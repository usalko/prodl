package sql_parser

import (
	"fmt"
	"testing"

	"github.com/usalko/prodl/internal/sql_parser"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
	"github.com/usalko/prodl/internal/sql_parser/psql"
)

func TestPsqlStatements(t *testing.T) {
	testcases := []struct {
		in string
		id []int
	}{
		{
			in: "-- comment\nCOMMENT ON SCHEMA public IS '1'",
			id: []int{psql.COMMENT, psql.ON, psql.SCHEMA, 0},
		},
		{
			in: "-- comment\n\nSET statement_timeout = 0",
			id: []int{psql.SET, 0},
		},
		{
			in: "\n\n--\n-- Name: public; Type: SCHEMA; Schema: -; Owner: phytonyms.dev\n--\n\n-- *not* creating schema, since initdb creates it\n\n\nALTER SCHEMA public OWNER TO \"phytonyms.dev\"",
			id: []int{psql.ALTER, 0},
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
