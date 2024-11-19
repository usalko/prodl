package archive_stream

import (
	"archive/zip"

	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/usalko/hexsi"
	"github.com/usalko/hexsi/ft"
)

var SUPPORTED_FORMATS map[ft.FileType]bool = map[ft.FileType]bool{
	ft.GZIP: false,
	ft.ZIP:  true,
}

type ArchiveEntryState struct {
	reader                     io.Reader
	limitedReader              io.Reader
	readNum                    uint64
	hasDataDescriptorSignature bool
	eof                        bool
}

func (entry *ArchiveEntryState) isHasDataDescriptorSignature() bool {
	return entry.hasDataDescriptorSignature
}

func (entry *ArchiveEntryState) getReadNum() uint64 {
	return entry.readNum
}

func (entry *ArchiveEntryState) setEof(eof bool) {
	entry.eof = eof
}

func (entry *ArchiveEntryState) isEof() bool {
	return entry.eof
}

func (entry *ArchiveEntryState) getUncompressedSize64() uint64 {
	return 0
}

func (entry *ArchiveEntryState) getLimitedReader() io.Reader {
	return entry.limitedReader
}

type ArchiveEntry interface {
	IsDir() bool
	Open() (io.ReadCloser, error)

	isHasDataDescriptorSignature() bool
	getReadNum() uint64
	setEof(eof bool)
	isEof() bool
	getUncompressedSize64() uint64
	getLimitedReader() io.Reader
	readDataDescriptor(r io.Reader) error

	addReadNum(n uint64)
	getReader() io.Reader
	getCrc32() uint32
	GetName() string
}

type ArchiveStreamReader struct {
	inputReader  io.Reader
	localFileEnd bool
	curEntry     ArchiveEntry
	fileType     ft.FileType
}

func NewReader(reader io.Reader) *ArchiveStreamReader {
	return &ArchiveStreamReader{
		inputReader: reader,
	}
}

func (reader *ArchiveStreamReader) readEntry() (ArchiveEntry, error) {

	buf := make([]byte, fileHeaderLen)
	if _, err := io.ReadFull(reader.inputReader, buf); err != nil {
		return nil, fmt.Errorf("unable to read local file header: %w", err)
	}

	lr := ReadBuf(buf)

	readerVersion := lr.Uint16()
	flags := lr.Uint16()
	method := lr.Uint16()
	modifiedTime := lr.Uint16()
	modifiedDate := lr.Uint16()
	crc32Sum := lr.Uint32()
	compressedSize := lr.Uint32()
	uncompressedSize := lr.Uint32()
	filenameLen := int(lr.Uint16())
	extraAreaLen := int(lr.Uint16())

	entry := &ZipEntry{
		FileHeader: zip.FileHeader{
			ReaderVersion:      readerVersion,
			Flags:              flags,
			Method:             method,
			ModifiedTime:       modifiedTime,
			ModifiedDate:       modifiedDate,
			CRC32:              crc32Sum,
			CompressedSize:     compressedSize,
			UncompressedSize:   uncompressedSize,
			CompressedSize64:   uint64(compressedSize),
			UncompressedSize64: uint64(uncompressedSize),
		},
		ArchiveEntryState: ArchiveEntryState{
			reader:  reader.inputReader,
			readNum: 0,
			eof:     false,
		},
	}

	nameAndExtraBuf := make([]byte, filenameLen+extraAreaLen)
	if _, err := io.ReadFull(reader.inputReader, nameAndExtraBuf); err != nil {
		return nil, fmt.Errorf("unable to read entry name and extra area: %w", err)
	}

	entry.Name = string(nameAndExtraBuf[:filenameLen])
	entry.Extra = nameAndExtraBuf[filenameLen:]

	entry.NonUTF8 = flags&0x800 == 0
	if flags&1 == 1 {
		return nil, fmt.Errorf("encrypted ZIP entry not supported")
	}
	if flags&8 == 8 && method != CompressMethodDeflated {
		return nil, fmt.Errorf("only DEFLATED entries can have data descriptor")
	}

	needCSize := entry.CompressedSize64 == ^uint64(0)
	needUSize := entry.UncompressedSize64 == ^uint64(0)

	ler := ReadBuf(entry.Extra)
	var modified time.Time
parseExtras:
	for len(ler) >= 4 { // need at least tag and size
		fieldTag := ler.Uint16()
		fieldSize := int(ler.Uint16())
		if len(ler) < fieldSize {
			break
		}
		fieldBuf := ler.Sub(fieldSize)

		switch fieldTag {
		case Zip64ExtraID:
			entry.zip64 = true

			// update directory values from the zip64 extra block.
			// They should only be consulted if the sizes read earlier
			// are maxed out.
			// See golang.org/issue/13367.
			if needUSize {
				needUSize = false
				if len(fieldBuf) < 8 {
					return nil, zip.ErrFormat
				}
				entry.UncompressedSize64 = fieldBuf.Uint64()
			}
			if needCSize {
				needCSize = false
				if len(fieldBuf) < 8 {
					return nil, zip.ErrFormat
				}
				entry.CompressedSize64 = fieldBuf.Uint64()
			}
		case NtFsExtraID:
			if len(fieldBuf) < 4 {
				continue parseExtras
			}
			fieldBuf.Uint32()        // reserved (ignored)
			for len(fieldBuf) >= 4 { // need at least tag and size
				attrTag := fieldBuf.Uint16()
				attrSize := int(fieldBuf.Uint16())
				if len(fieldBuf) < attrSize {
					continue parseExtras
				}
				attrBuf := fieldBuf.Sub(attrSize)
				if attrTag != 1 || attrSize != 24 {
					continue // Ignore irrelevant attributes
				}

				const ticksPerSecond = 1e7    // Windows timestamp resolution
				ts := int64(attrBuf.Uint64()) // ModTime since Windows epoch
				secs := ts / ticksPerSecond
				nsecs := (1e9 / ticksPerSecond) * int64(ts%ticksPerSecond)
				epoch := time.Date(1601, time.January, 1, 0, 0, 0, 0, time.UTC)
				modified = time.Unix(epoch.Unix()+secs, nsecs)
			}
		case UnixExtraID, InfoZipUnixExtraID:
			if len(fieldBuf) < 8 {
				continue parseExtras
			}
			fieldBuf.Uint32()              // AcTime (ignored)
			ts := int64(fieldBuf.Uint32()) // ModTime since Unix epoch
			modified = time.Unix(ts, 0)
		case ExtTimeExtraID:
			if len(fieldBuf) < 5 || fieldBuf.Uint8()&1 == 0 {
				continue parseExtras
			}
			ts := int64(fieldBuf.Uint32()) // ModTime since Unix epoch
			modified = time.Unix(ts, 0)
		}
	}

	msDosModified := MSDosTimeToTime(entry.ModifiedDate, entry.ModifiedTime)
	entry.Modified = msDosModified

	if !modified.IsZero() {
		entry.Modified = modified.UTC()

		// If legacy MS-DOS timestamps are set, we can use the delta between
		// the legacy and extended versions to estimate timezone offset.
		//
		// A non-UTC timezone is always used (even if offset is zero).
		// Thus, FileHeader.Modified.Location() == time.UTC is useful for
		// determining whether extended timestamps are present.
		// This is necessary for users that need to do additional time
		// calculations when dealing with legacy ZIP formats.
		if entry.ModifiedDate != 0 || entry.ModifiedTime != 0 {
			entry.Modified = modified.In(timeZone(msDosModified.Sub(modified)))
		}
	}

	if needCSize {
		return nil, zip.ErrFormat
	}

	entry.limitedReader = io.LimitReader(reader.inputReader, int64(entry.CompressedSize64))

	return entry, nil
}

func (reader *ArchiveStreamReader) GetNextEntry() (ArchiveEntry, error) {
	if reader.localFileEnd {
		return nil, io.EOF
	}
	if reader.curEntry != nil && !reader.curEntry.isEof() {
		if reader.curEntry.getReadNum() <= reader.curEntry.getUncompressedSize64() {
			if _, err := io.Copy(io.Discard, reader.curEntry.getLimitedReader()); err != nil {
				return nil, fmt.Errorf("read previous file data fail: %w", err)
			}
			if reader.curEntry.isHasDataDescriptorSignature() {
				if err := reader.curEntry.readDataDescriptor(reader.inputReader); err != nil {
					return nil, fmt.Errorf("read previous entry's data descriptor fail: %w", err)
				}
			}
		} else {
			if !reader.curEntry.isHasDataDescriptorSignature() {
				return nil, errors.New("parse error, read position exceed entry")
			}

			readDataLen := reader.curEntry.getReadNum() - reader.curEntry.getUncompressedSize64()
			if readDataLen > dataDescriptorLen {
				return nil, errors.New("parse error, read position exceed entry")
			} else if readDataLen > dataDescriptorLen-4 {
				if reader.curEntry.isHasDataDescriptorSignature() {
					if _, err := io.Copy(io.Discard, io.LimitReader(reader.inputReader, int64(dataDescriptorLen-readDataLen))); err != nil {
						return nil, fmt.Errorf("read previous entry's data descriptor fail: %w", err)
					}
				} else {
					return nil, errors.New("parse error, read position exceed entry")
				}
			} else {
				buf := make([]byte, dataDescriptorLen-readDataLen)
				if _, err := io.ReadFull(reader.inputReader, buf); err != nil {
					return nil, fmt.Errorf("read previous entry's data descriptor fail: %w", err)
				}
				buf = buf[len(buf)-4:]
				headerID := binary.LittleEndian.Uint32(buf)

				// read to next record head
				if headerID == fileHeaderSignature ||
					headerID == directoryHeaderSignature ||
					headerID == directoryEndSignature {
					reader.inputReader = io.MultiReader(bytes.NewReader(buf), reader.inputReader)
				}
			}
		}
		reader.curEntry.setEof(true)
	}
	headerIDBuf := make([]byte, headerIdentifierLen)
	if _, err := io.ReadFull(reader.inputReader, headerIDBuf); err != nil {
		return nil, fmt.Errorf("unable to read header identifier: %w", err)
	}

	fileType, _ := hexsi.DetectFileType(headerIDBuf)
	if fileType != nil {
		reader.fileType = *fileType
		if _, ok := SUPPORTED_FORMATS[*fileType]; !ok {
			return nil, fmt.Errorf("unsupported archive format")
		}
	} else {
		headerID := binary.LittleEndian.Uint32(headerIDBuf)
		if headerID == directoryHeaderSignature || headerID == directoryEndSignature {
			reader.localFileEnd = true
			return nil, io.EOF
		}
		return nil, zip.ErrFormat
	}

	entry, err := reader.readEntry()
	if err != nil {
		return nil, fmt.Errorf("unable to read file header: %w", err)
	}
	reader.curEntry = entry
	return entry, nil
}