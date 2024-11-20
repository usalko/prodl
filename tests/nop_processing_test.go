package tests

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"github.com/usalko/sent/internal/archive_stream"
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

func TestNopProcessing(t *testing.T) {
	fileName := "test_data/test_case1.txt.gz"
	respBody, err := os.Open(fileName)
	check(err, "File %s open error", fileName)

	reader := archive_stream.NewReader(respBody)

	for {
		entry, err := reader.GetNextEntry()
		if err == io.EOF {
			break
		}
		check(err, "unable to get next entry")

		// log.Println("entry name: ", entry.Name)
		// log.Println("entry comment: ", entry.Comment)
		// log.Println("entry reader version: ", entry.ReaderVersion)
		// log.Println("entry modify time: ", entry.Modified)
		// log.Println("entry compressed size: ", entry.CompressedSize64)
		// log.Println("entry uncompressed size: ", entry.UncompressedSize64)
		log.Println("entry is a dir: ", entry.IsDir())

		if !entry.IsDir() {
			rc, err := entry.Open()
			if err != nil {
				log.Fatalf("unable to open gzip file: %s", err)
			}
			content, err := io.ReadAll(rc)
			if err != nil {
				log.Fatalf("read gzip file content fail: %s", err)
			}

			log.Println("file length:", len(content))

			// if uint64(len(content)) != entry.UncompressedSize64 {
			// 	log.Fatalf("read zip file length not equal with UncompressedSize64")
			// }
			if err := rc.Close(); err != nil {
				log.Fatalf("close gzip entry reader fail: %s", err)
			}
		}
	}
}
