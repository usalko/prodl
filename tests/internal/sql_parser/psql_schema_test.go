package sql_parser

import (
	"fmt"
	"testing"

	"github.com/usalko/prodl/internal/sql_parser"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
	"github.com/usalko/prodl/internal/sql_parser/psql"
)

func TestCommentOnSchema(t *testing.T) {
	testcases := []struct {
		in string
		id []int
	}{
		{
			in: "-- comment\nCOMMENT ON SCHEMA public IS '1'",
			// in: "SET a = '1'",
			id: []int{psql.COMMENT, psql.ON, psql.SCHEMA, 0},
		},
	}

	for _, tcase := range testcases {
		t.Run(tcase.in, func(t *testing.T) {
			tok, err := sql_parser.Parse(tcase.in, dialect.PSQL)
			if err != nil {
				t.Fatalf("%q", err)
			}
			fmt.Printf("tok: %v\n", tok)
		})
	}
}
