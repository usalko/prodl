package sql_parser

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/usalko/prodl/internal/sql_parser/tokenizer"
)

func TestCyclicBufferCase1(t *testing.T) {
	fmt.Printf("%T\n", 'w')
	fmt.Println(reflect.TypeOf('w').String())
	var r rune = 'w'
	fmt.Printf("%T\n", r)

	text := "首映鼓掌10分鐘"
	fmt.Printf("String (len: %d) element type is %T\n", len(text), text[1])
	for _, r := range "string" {
		fmt.Printf("Range string element type is %T\n", r)
		break
	}

	cBuf := tokenizer.NewCyclicBuffer(5)
	cBuf.Put('a', 'b', 'c', 'd', 'e', 'f', 'g')
	if !cBuf.Has("cd", "fg") {
		t.Errorf("wrong buffer content %s", cBuf)
	}
}

func TestCyclicBufferCase2(t *testing.T) {
	cBuf := tokenizer.NewCyclicBuffer(5)
	cBuf.Put('a', 'b', 'c', 'd', 'e', 'f', 'g', 'a', 'b', 'c', 'd', 'e', 'f', 'g')
	if !cBuf.Has("cd", "fg") {
		t.Errorf("wrong buffer content %s", cBuf)
	}
}
