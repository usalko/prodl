package archive_stream

import (
	"compress/flate"
	"errors"
	"io"
	"sync"
)

var flateReaderPool sync.Pool

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

func newFlateReader(r io.Reader) io.ReadCloser {
	fr, ok := flateReaderPool.Get().(io.ReadCloser)
	if ok {
		fr.(flate.Resetter).Reset(r, nil)
	} else {
		fr = flate.NewReader(r)
	}
	return &pooledFlateReader{fr: fr}
}
