package archive_stream

import (
	"archive/zip"
	"errors"
	"hash/crc32"
	"io"
	"sync"
)

type ZipEntry struct {
	zip.FileHeader // Entry header
	ArchiveEntryState
	zip64 bool
}

// GetName implements ArchiveEntry.
func (entry *ZipEntry) GetName() string {
	return entry.Name
}

// IsDir implements ArchiveEntry.
// Just simply check whether the entry name ends with "/"
func (entry *ZipEntry) IsDir() bool {
	return len(entry.Name) > 0 && entry.Name[len(entry.Name)-1] == '/'
}

// Open implements ArchiveEntry.
func (entry *ZipEntry) Open() (io.ReadCloser, error) {
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

// addReadNum implements ArchiveEntry.
func (entry *ZipEntry) addReadNum(n uint64) {
	entry.readNum += n
}

// getCrc32 implements ArchiveEntry.
func (entry *ZipEntry) getCrc32() uint32 {
	return entry.CRC32
}

// getLimitedReader implements ArchiveEntry.
func (entry *ZipEntry) getLimitedReader() io.Reader {
	return entry.limitedReader
}

// getReadNum implements ArchiveEntry.
func (entry *ZipEntry) getReadNum() uint64 {
	return entry.readNum
}

// getReader implements ArchiveEntry.
func (entry *ZipEntry) getReader() io.Reader {
	return entry.reader
}

// getUncompressedSize64 implements ArchiveEntry.
func (entry *ZipEntry) getUncompressedSize64() uint64 {
	return entry.CompressedSize64
}

// isEof implements ArchiveEntry.
func (entry *ZipEntry) isEof() bool {
	return entry.eof
}

// isHasDataDescriptorSignature implements ArchiveEntry.
func (entry *ZipEntry) isHasDataDescriptorSignature() bool {
	return entry.Flags&8 != 0
}

// readDataDescriptor implements ArchiveEntry.
func (entry *ZipEntry) readDataDescriptor(r io.Reader) error {
	var buf [sipDataDescriptorLen]byte
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
	entry.readNum += uint64(n)
	if err != nil {
		return err
	}
	off := 0
	maybeSig := ReadBuf(buf[:4])
	if maybeSig.Uint32() != zipDataDescriptorSignature {
		// No data descriptor signature. Keep these four
		// bytes.
		off += 4
	} else {
		entry.hasDataDescriptorSignature = true
	}
	n, err = io.ReadFull(r, buf[off:12])
	entry.readNum += uint64(n)
	if err != nil {
		return err
	}
	(*entry).eof = true
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

// setEof implements ArchiveEntry.
func (entry *ZipEntry) setEof(eof bool) {
	entry.eof = eof
}

const (
	// Zip files signatures
	zipHeaderIdentifierLen      = 4
	zipFileHeaderLen            = 26
	sipDataDescriptorLen        = 16 // four uint32: descriptor signature, crc32, compressed size, size
	zipFileHeaderSignature      = 0x04034b50
	zipDirectoryHeaderSignature = 0x02014b50
	zipDirectoryEndSignature    = 0x06054b50
	zipDataDescriptorSignature  = 0x08074b50

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
