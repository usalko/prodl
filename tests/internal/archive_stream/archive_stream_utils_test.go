package archive_stream_tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/usalko/prodl/internal/archive_stream"
)

func TestReadBuf(t *testing.T) {

	buf := []byte{
		0x01,
		0x02, 0x03,
		0x4, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	}

	lb := archive_stream.ReadBuf(buf)

	if !strings.HasSuffix("01", fmt.Sprintf("%x", lb.Uint8())) {
		t.Fatalf("uint8 read err")
	}

	if !strings.HasSuffix("0302", fmt.Sprintf("%x", lb.Uint16())) {
		t.Fatalf("uint16 read err")
	}

	if !strings.HasSuffix("07060504", fmt.Sprintf("%x", lb.Uint32())) {
		t.Fatalf("uint32 read err")
	}

	if !strings.HasSuffix("0f0e0d0c0b0a0908", fmt.Sprintf("%x", lb.Uint64())) {
		t.Fatalf("uint64 read err")
	}

}
