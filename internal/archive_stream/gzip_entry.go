package archive_stream

import (
	"compress/gzip"
	"io"
)

type GzipEntry struct {
	gzip.Header // Entry Header
	ArchiveEntryState
}

type GzipEntryCloser struct {
	io.Reader
	gzipEntry *GzipEntry
}

func (gzipEntryCloser GzipEntryCloser) Close() error {
	gzipEntryCloser.gzipEntry.eof = true
	return nil
}

// GetName implements ArchiveEntry.
func (entry *GzipEntry) GetName() string {
	return entry.Header.Name
}

// IsDir implements ArchiveEntry.
func (entry *GzipEntry) IsDir() bool {
	return false
}

// Open implements ArchiveEntry.
func (entry *GzipEntry) Open() (io.ReadCloser, error) {
	return GzipEntryCloser{
		Reader:    entry.reader,
		gzipEntry: entry,
	}, nil
}

// addReadNum implements ArchiveEntry.
func (entry *GzipEntry) addReadNum(n uint64) {
	entry.readNum += n
}

// getCrc32 implements ArchiveEntry.
func (entry *GzipEntry) getCrc32() uint32 {
	return 0
}

// getLimitedReader implements ArchiveEntry.
func (entry *GzipEntry) getLimitedReader() io.Reader {
	return io.LimitReader(entry.reader, 0)
}

// getReadNum implements ArchiveEntry.
func (entry *GzipEntry) getReadNum() uint64 {
	return entry.readNum
}

// getReader implements ArchiveEntry.
func (entry *GzipEntry) getReader() io.Reader {
	return entry.reader
}

// getUncompressedSize64 implements ArchiveEntry.
func (entry *GzipEntry) getUncompressedSize64() uint64 {
	return 0
}

// isEof implements ArchiveEntry.
func (entry *GzipEntry) isEof() bool {
	return entry.eof
}

// isHasDataDescriptorSignature implements ArchiveEntry.
func (entry *GzipEntry) isHasDataDescriptorSignature() bool {
	return true
}

// readDataDescriptor implements ArchiveEntry.
func (entry *GzipEntry) readDataDescriptor(r io.Reader) error {
	return nil
}

// setEof implements ArchiveEntry.
func (entry *GzipEntry) setEof(eof bool) {
	entry.eof = eof
}
