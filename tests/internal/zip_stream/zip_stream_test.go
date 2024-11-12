package zip_stream

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"github.com/usalko/sent/internal/zip_stream"
)

func check(err error, msgs ...any) {
	if err != nil {
		if len(msgs) == 0 {
			panic(err)
		} else if len(msgs) == 1 {
			panic(fmt.Errorf("%s: %s", msgs[0], err))
		} else {
			panic(fmt.Errorf("%s: %s", fmt.Sprintf(msgs[0].(string), msgs[1:]...), err))
		}
	}
}

func TestStreamReader(t *testing.T) {
	fileName := "test_data/testing.txt.gz"
	respBody, err := os.Open(fileName)
	check(err, "File %s open error", fileName)

	reader := zip_stream.NewReader(respBody)

	for {
		entry, err := reader.GetNextEntry()
		if err == io.EOF {
			break
		}
		check(err, "unable to get next entry")

		log.Println("entry name: ", entry.Name)
		log.Println("entry comment: ", entry.Comment)
		log.Println("entry reader version: ", entry.ReaderVersion)
		log.Println("entry modify time: ", entry.Modified)
		log.Println("entry compressed size: ", entry.CompressedSize64)
		log.Println("entry uncompressed size: ", entry.UncompressedSize64)
		log.Println("entry is a dir: ", entry.IsDir())

		if !entry.IsDir() {
			rc, err := entry.Open()
			if err != nil {
				log.Fatalf("unable to open zip file: %s", err)
			}
			content, err := io.ReadAll(rc)
			if err != nil {
				log.Fatalf("read zip file content fail: %s", err)
			}

			log.Println("file length:", len(content))

			if uint64(len(content)) != entry.UncompressedSize64 {
				log.Fatalf("read zip file length not equal with UncompressedSize64")
			}
			if err := rc.Close(); err != nil {
				log.Fatalf("close zip entry reader fail: %s", err)
			}
		}
	}
}

func TestNewReader(t *testing.T) {

	f, err := os.Open("test_data/testing.txt.zip")
	check(err)
	defer f.Close()

	zipFile, err := os.ReadFile("test_data/testing.txt.zip")
	check(err)

	az, err := zip.NewReader(f, int64(len(zipFile)))
	check(err)

	fileMap := make(map[string]*zip.File, len(az.File))

	for _, zf := range az.File {
		fileMap[zf.Name] = zf
	}

	z := zip_stream.NewReader(f)

	for {
		entry, err := z.GetNextEntry()
		if err == io.EOF {
			// iterator over
			break
		}

		zf, ok := fileMap[entry.Name]
		if !ok {
			t.Fatalf("not expected file: %s", entry.Name)
		}
		delete(fileMap, entry.Name)

		if entry.Comment != zf.Comment ||
			entry.ReaderVersion != zf.ReaderVersion ||
			entry.IsDir() != zf.Mode().IsDir() ||
			entry.Flags != zf.Flags ||
			entry.Method != zf.Method ||
			!entry.Modified.Equal(zf.Modified) ||
			entry.CRC32 != zf.CRC32 ||
			//bytes.Compare(entry.Extra, zf.Extra) != 0 || // local file header's extra data may not same as central directory header's extra data
			entry.CompressedSize64 != zf.CompressedSize64 ||
			entry.UncompressedSize64 != zf.UncompressedSize64 {
			t.Fatal("some local file header attr is incorrect")
		}

		if !entry.IsDir() {
			rc, err := entry.Open()
			if err != nil {
				t.Fatalf("open zip file entry err: %s", err)
			}

			entryFileContents, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("read entry file contents fail: %s", err)
			}

			ziprc, err := zf.Open()
			if err != nil {
				t.Fatal(err)
			}
			zipFileContents, err := io.ReadAll(ziprc)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(entryFileContents, zipFileContents) {
				t.Fatal("the zip entry file contents is incorrect")
			}

			if err := rc.Close(); err != nil {
				t.Fatalf("close zip file entry reader err: %s", err)
			}
			_ = ziprc.Close()
		}
	}

	if len(fileMap) != 0 {
		t.Fatal("the resolved entry count is incorrect")
	}

}
