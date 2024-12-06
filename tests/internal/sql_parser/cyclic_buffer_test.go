package sql_parser

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/usalko/prodl/internal/sql_parser/tokenizer"
)

func TestCyclicBuffer(t *testing.T) {
	fmt.Printf("%T\n", 'w')
	fmt.Println(reflect.TypeOf('w').String())
	var r rune = 'w'
	fmt.Printf("%T\n", r)

	tokenizer.NewCyclicBuffer(5)
}
