package sql_parser

import (
	"fmt"
	"testing"

	"github.com/usalko/prodl/internal/sql_parser"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
	"github.com/usalko/prodl/internal/sql_parser/sqlite3"
)

func TestSqlite3Statements(t *testing.T) {
	testcases := []struct {
		in string
		id []int
	}{
		{
			in: "SELECT * FROM my_table",
			// in: "SET a = '1'",
			id: []int{sqlite3.SELECT, 0},
		},
		{
			in: "SELECT column2, column1 FROM my_table",
			id: []int{sqlite3.SELECT, 0},
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
