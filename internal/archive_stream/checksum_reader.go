package archive_stream

import (
	"archive/zip"
	"hash"
	"io"
)

type checksumReader struct {
	rc    io.ReadCloser
	hash  hash.Hash32
	nread uint64 // number of bytes read so far
	entry ArchiveEntry
	err   error // sticky error
}

func (reader *checksumReader) Read(buff []byte) (n int, err error) {
	if reader.err != nil {
		return 0, reader.err
	}
	n, err = reader.rc.Read(buff)
	reader.hash.Write(buff[:n])
	reader.nread += uint64(n)
	reader.entry.addReadNum(uint64(n))
	if err == nil {
		return
	}
	if err == io.EOF {
		if reader.nread != reader.entry.getUncompressedSize64() {
			return 0, io.ErrUnexpectedEOF
		}
		if reader.entry.isHasDataDescriptorSignature() {
			if err1 := reader.entry.readDataDescriptor(reader.entry.getReader()); err1 != nil {
				if err1 == io.EOF {
					err = io.ErrUnexpectedEOF
				} else {
					err = err1
				}
			} else if reader.hash.Sum32() != reader.entry.getCrc32() {
				err = zip.ErrChecksum
			}
		} else {
			// If there's not a data descriptor, we still compare
			// the CRC32 of what we've read against the file header
			// or TOC's CRC32, if it seems like it was set.
			reader.entry.setEof(true)
			if reader.entry.getCrc32() != 0 && reader.hash.Sum32() != reader.entry.getCrc32() {
				err = zip.ErrChecksum
			}
		}
	}
	reader.err = err
	return
}

func (r *checksumReader) Close() error { return r.rc.Close() }
