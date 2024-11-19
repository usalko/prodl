package archive_stream

import "archive/tar"

type TarEntry struct {
	tar.Header // Entry Header
	ArchiveEntryState
}
