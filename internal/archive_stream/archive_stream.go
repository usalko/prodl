package archive_stream

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"sync"
	"time"

	"github.com/usalko/hexsi"
	"github.com/usalko/hexsi/ft"
)

const (
	// Zip files signatures
	headerIdentifierLen      = 4
	fileHeaderLen            = 26
	dataDescriptorLen        = 16 // four uint32: descriptor signature, crc32, compressed size, size
	fileHeaderSignature      = 0x04034b50
	directoryHeaderSignature = 0x02014b50
	directoryEndSignature    = 0x06054b50
	dataDescriptorSignature  = 0x08074b50

	// Extra header IDs.
	// See http://mdfs.net/Docs/Comp/Archiving/Zip/ExtraField
	Zip64ExtraID       = 0x0001 // Zip64 extended information
	NtFsExtraID        = 0x000a // NTFS
	UnixExtraID        = 0x000d // UNIX
	ExtTimeExtraID     = 0x5455 // Extended timestamp
	InfoZipUnixExtraID = 0x5855 // Info-ZIP Unix extension

)

const (
	CompressMethodStored   = 0
	CompressMethodDeflated = 8
)

type Entry struct {
	zip.FileHeader
	reader                     io.Reader
	limitedReader              io.Reader
	zip64                      bool
	hasReadNum                 uint64
	hasDataDescriptorSignature bool
	eof                        bool
}

func (entry *Entry) hasDataDescriptor() bool {
	return entry.Flags&8 != 0
}

// IsDir just simply check whether the entry name ends with "/"
func (entry *Entry) IsDir() bool {
	return len(entry.Name) > 0 && entry.Name[len(entry.Name)-1] == '/'
}

func (entry *Entry) Open() (io.ReadCloser, error) {
	if entry.eof {
		return nil, errors.New("this file has read to end")
	}
	decomp := decompressor(entry.Method)
	if decomp == nil {
		return nil, zip.ErrAlgorithm
	}
	rc := decomp(entry.limitedReader)

	return &checksumReader{
		rc:    rc,
		hash:  crc32.NewIEEE(),
		entry: entry,
	}, nil
}

type ArchiveStreamReader struct {
	reader       io.Reader
	localFileEnd bool
	curEntry     *Entry
}

func NewReader(_reader io.Reader) *ArchiveStreamReader {
	return &ArchiveStreamReader{
		reader: _reader,
	}
}

func (reader *ArchiveStreamReader) readEntry() (*Entry, error) {

	buf := make([]byte, fileHeaderLen)
	if _, err := io.ReadFull(reader.reader, buf); err != nil {
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

	entry := &Entry{
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
		reader:     reader.reader,
		hasReadNum: 0,
		eof:        false,
	}

	nameAndExtraBuf := make([]byte, filenameLen+extraAreaLen)
	if _, err := io.ReadFull(reader.reader, nameAndExtraBuf); err != nil {
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

	entry.limitedReader = io.LimitReader(reader.reader, int64(entry.CompressedSize64))

	return entry, nil
}

func (reader *ArchiveStreamReader) GetNextEntry() (*Entry, error) {
	if reader.localFileEnd {
		return nil, io.EOF
	}
	if reader.curEntry != nil && !reader.curEntry.eof {
		if reader.curEntry.hasReadNum <= reader.curEntry.UncompressedSize64 {
			if _, err := io.Copy(io.Discard, reader.curEntry.limitedReader); err != nil {
				return nil, fmt.Errorf("read previous file data fail: %w", err)
			}
			if reader.curEntry.hasDataDescriptor() {
				if err := readDataDescriptor(reader.reader, reader.curEntry); err != nil {
					return nil, fmt.Errorf("read previous entry's data descriptor fail: %w", err)
				}
			}
		} else {
			if !reader.curEntry.hasDataDescriptor() {
				return nil, errors.New("parse error, read position exceed entry")
			}

			readDataLen := reader.curEntry.hasReadNum - reader.curEntry.UncompressedSize64
			if readDataLen > dataDescriptorLen {
				return nil, errors.New("parse error, read position exceed entry")
			} else if readDataLen > dataDescriptorLen-4 {
				if reader.curEntry.hasDataDescriptorSignature {
					if _, err := io.Copy(io.Discard, io.LimitReader(reader.reader, int64(dataDescriptorLen-readDataLen))); err != nil {
						return nil, fmt.Errorf("read previous entry's data descriptor fail: %w", err)
					}
				} else {
					return nil, errors.New("parse error, read position exceed entry")
				}
			} else {
				buf := make([]byte, dataDescriptorLen-readDataLen)
				if _, err := io.ReadFull(reader.reader, buf); err != nil {
					return nil, fmt.Errorf("read previous entry's data descriptor fail: %w", err)
				}
				buf = buf[len(buf)-4:]
				headerID := binary.LittleEndian.Uint32(buf)

				// read to next record head
				if headerID == fileHeaderSignature ||
					headerID == directoryHeaderSignature ||
					headerID == directoryEndSignature {
					reader.reader = io.MultiReader(bytes.NewReader(buf), reader.reader)
				}
			}
		}
		reader.curEntry.eof = true
	}
	headerIDBuf := make([]byte, headerIdentifierLen)
	if _, err := io.ReadFull(reader.reader, headerIDBuf); err != nil {
		return nil, fmt.Errorf("unable to read header identifier: %w", err)
	}

	fileType, _ := hexsi.DetectFileType(headerIDBuf)
	if *fileType == ft.GZIP {
		panic(fmt.Errorf("unsupported archive format"))
	}

	headerID := binary.LittleEndian.Uint32(headerIDBuf)
	if headerID != fileHeaderSignature {
		if headerID == directoryHeaderSignature || headerID == directoryEndSignature {
			reader.localFileEnd = true
			return nil, io.EOF
		}
		return nil, zip.ErrFormat
	}
	entry, err := reader.readEntry()
	if err != nil {
		return nil, fmt.Errorf("unable to read zip file header: %w", err)
	}
	reader.curEntry = entry
	return entry, nil
}

var (
	decompressors sync.Map // map[uint16]Decompressor
)

func init() {
	decompressors.Store(zip.Store, zip.Decompressor(io.NopCloser))
	decompressors.Store(zip.Deflate, zip.Decompressor(newFlateReader))
}

func decompressor(method uint16) zip.Decompressor {
	di, ok := decompressors.Load(method)
	if !ok {
		return nil
	}
	return di.(zip.Decompressor)
}

var flateReaderPool sync.Pool

func newFlateReader(r io.Reader) io.ReadCloser {
	fr, ok := flateReaderPool.Get().(io.ReadCloser)
	if ok {
		fr.(flate.Resetter).Reset(r, nil)
	} else {
		fr = flate.NewReader(r)
	}
	return &pooledFlateReader{fr: fr}
}

type pooledFlateReader struct {
	mu sync.Mutex // guards Close and Read
	fr io.ReadCloser
}

func (reader *pooledFlateReader) Read(p []byte) (n int, err error) {
	reader.mu.Lock()
	defer reader.mu.Unlock()
	if reader.fr == nil {
		return 0, errors.New("Read after Close")
	}
	return reader.fr.Read(p)
}

func (reader *pooledFlateReader) Close() error {
	reader.mu.Lock()
	defer reader.mu.Unlock()
	var err error
	if reader.fr != nil {
		err = reader.fr.Close()
		flateReaderPool.Put(reader.fr)
		reader.fr = nil
	}
	return err
}

func readDataDescriptor(r io.Reader, entry *Entry) error {
	var buf [dataDescriptorLen]byte
	// From the spec: "Although not originally assigned a
	// signature, the value 0x08074b50 has commonly been adopted
	// as a signature value for the data descriptor record.
	// Implementers should be aware that ZIP files may be
	// encountered with or without this signature marking data
	// descriptors and should account for either case when reading
	// ZIP files to ensure compatibility."
	//
	// dataDescriptorLen includes the size of the signature but
	// first read just those 4 bytes to see if it exists.
	n, err := io.ReadFull(r, buf[:4])
	entry.hasReadNum += uint64(n)
	if err != nil {
		return err
	}
	off := 0
	maybeSig := ReadBuf(buf[:4])
	if maybeSig.Uint32() != dataDescriptorSignature {
		// No data descriptor signature. Keep these four
		// bytes.
		off += 4
	} else {
		entry.hasDataDescriptorSignature = true
	}
	n, err = io.ReadFull(r, buf[off:12])
	entry.hasReadNum += uint64(n)
	if err != nil {
		return err
	}
	entry.eof = true
	b := ReadBuf(buf[:12])
	if b.Uint32() != entry.CRC32 {
		return zip.ErrChecksum
	}

	// The two sizes that follow here can be either 32 bits or 64 bits
	// but the spec is not very clear on this and different
	// interpretations has been made causing incompatibilities. We
	// already have the sizes from the central directory so we can
	// just ignore these.

	return nil
}

type checksumReader struct {
	rc    io.ReadCloser
	hash  hash.Hash32
	nread uint64 // number of bytes read so far
	entry *Entry
	err   error // sticky error
}

func (reader *checksumReader) Read(buff []byte) (n int, err error) {
	if reader.err != nil {
		return 0, reader.err
	}
	n, err = reader.rc.Read(buff)
	reader.hash.Write(buff[:n])
	reader.nread += uint64(n)
	reader.entry.hasReadNum += uint64(n)
	if err == nil {
		return
	}
	if err == io.EOF {
		if reader.nread != reader.entry.UncompressedSize64 {
			return 0, io.ErrUnexpectedEOF
		}
		if reader.entry.hasDataDescriptor() {
			if err1 := readDataDescriptor(reader.entry.reader, reader.entry); err1 != nil {
				if err1 == io.EOF {
					err = io.ErrUnexpectedEOF
				} else {
					err = err1
				}
			} else if reader.hash.Sum32() != reader.entry.CRC32 {
				err = zip.ErrChecksum
			}
		} else {
			// If there's not a data descriptor, we still compare
			// the CRC32 of what we've read against the file header
			// or TOC's CRC32, if it seems like it was set.
			reader.entry.eof = true
			if reader.entry.CRC32 != 0 && reader.hash.Sum32() != reader.entry.CRC32 {
				err = zip.ErrChecksum
			}
		}
	}
	reader.err = err
	return
}

func (r *checksumReader) Close() error { return r.rc.Close() }
