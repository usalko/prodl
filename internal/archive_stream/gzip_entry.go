package archive_stream

import (
	"compress/gzip"
	"io"
)

type GzipEntry struct {
	gzip.Header // Entry Header
	ArchiveEntryState
}

// GetName implements ArchiveEntry.
func (g *GzipEntry) GetName() string {
	return g.Header.Name
}

// IsDir implements ArchiveEntry.
func (g *GzipEntry) IsDir() bool {
	return false
}

// Open implements ArchiveEntry.
func (g *GzipEntry) Open() (io.ReadCloser, error) {
	panic("unimplemented")
}

// addReadNum implements ArchiveEntry.
func (g *GzipEntry) addReadNum(n uint64) {
	panic("unimplemented")
}

// getCrc32 implements ArchiveEntry.
func (g *GzipEntry) getCrc32() uint32 {
	panic("unimplemented")
}

// getLimitedReader implements ArchiveEntry.
func (g *GzipEntry) getLimitedReader() io.Reader {
	panic("unimplemented")
}

// getReadNum implements ArchiveEntry.
func (g *GzipEntry) getReadNum() uint64 {
	panic("unimplemented")
}

// getReader implements ArchiveEntry.
func (g *GzipEntry) getReader() io.Reader {
	panic("unimplemented")
}

// getUncompressedSize64 implements ArchiveEntry.
func (g *GzipEntry) getUncompressedSize64() uint64 {
	panic("unimplemented")
}

// isEof implements ArchiveEntry.
func (g *GzipEntry) isEof() bool {
	panic("unimplemented")
}

// isHasDataDescriptorSignature implements ArchiveEntry.
func (g *GzipEntry) isHasDataDescriptorSignature() bool {
	panic("unimplemented")
}

// readDataDescriptor implements ArchiveEntry.
func (g *GzipEntry) readDataDescriptor(r io.Reader) error {
	panic("unimplemented")
}

// setEof implements ArchiveEntry.
func (g *GzipEntry) setEof(eof bool) {
	panic("unimplemented")
}
