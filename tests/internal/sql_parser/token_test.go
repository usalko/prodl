/*
Copyright 2019 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sql_parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/usalko/prodl/internal/sql_parser"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
	"github.com/usalko/prodl/internal/sql_parser/mysql"
)

func TestLiteralID(t *testing.T) {
	testcases := []struct {
		in  string
		id  int
		out string
	}{{
		in:  "`aa`",
		id:  mysql.ID,
		out: "aa",
	}, {
		in:  "```a```",
		id:  mysql.ID,
		out: "`a`",
	}, {
		in:  "`a``b`",
		id:  mysql.ID,
		out: "a`b",
	}, {
		in:  "`a``b`c",
		id:  mysql.ID,
		out: "a`b",
	}, {
		in:  "`a``b",
		id:  mysql.LEX_ERROR,
		out: "a`b",
	}, {
		in:  "`a``b``",
		id:  mysql.LEX_ERROR,
		out: "a`b`",
	}, {
		in:  "``",
		id:  mysql.LEX_ERROR,
		out: "",
	}, {
		in:  "@x",
		id:  mysql.AT_ID,
		out: "x",
	}, {
		in:  "@@x",
		id:  mysql.AT_AT_ID,
		out: "x",
	}, {
		in:  "@@`x y`",
		id:  mysql.AT_AT_ID,
		out: "x y",
	}, {
		in:  "@@`@x @y`",
		id:  mysql.AT_AT_ID,
		out: "@x @y",
	}}

	for _, tcase := range testcases {
		t.Run(tcase.in, func(t *testing.T) {
			tkn, err := sql_parser.NewStringTokenizer(tcase.in, dialect.MYSQL)
			if err != nil {
				t.Fatalf("%q", err)
			}
			id, out := tkn.Scan()
			require.Equal(t, tcase.id, id)
			require.Equal(t, tcase.out, string(out))
		})
	}
}

func tokenName(id int) string {
	if id == mysql.STRING {
		return "STRING"
	} else if id == mysql.LEX_ERROR {
		return "LEX_ERROR"
	}
	return fmt.Sprintf("%d", id)
}

func TestString(t *testing.T) {
	testcases := []struct {
		in   string
		id   int
		want string
	}{{
		in:   "''",
		id:   mysql.STRING,
		want: "",
	}, {
		in:   "''''",
		id:   mysql.STRING,
		want: "'",
	}, {
		in:   "'hello'",
		id:   mysql.STRING,
		want: "hello",
	}, {
		in:   "'\\n'",
		id:   mysql.STRING,
		want: "\n",
	}, {
		in:   "'\\nhello\\n'",
		id:   mysql.STRING,
		want: "\nhello\n",
	}, {
		in:   "'a''b'",
		id:   mysql.STRING,
		want: "a'b",
	}, {
		in:   "'a\\'b'",
		id:   mysql.STRING,
		want: "a'b",
	}, {
		in:   "'\\'",
		id:   mysql.LEX_ERROR,
		want: "'",
	}, {
		in:   "'",
		id:   mysql.LEX_ERROR,
		want: "",
	}, {
		in:   "'hello\\'",
		id:   mysql.LEX_ERROR,
		want: "hello'",
	}, {
		in:   "'hello",
		id:   mysql.LEX_ERROR,
		want: "hello",
	}, {
		in:   "'hello\\",
		id:   mysql.LEX_ERROR,
		want: "hello",
	}}

	for _, tcase := range testcases {
		t.Run(tcase.in, func(t *testing.T) {
			tokenizer, err := sql_parser.NewStringTokenizer(tcase.in, dialect.MYSQL)
			if err != nil {
				t.Fatalf("%q", err)
			}
			id, got := tokenizer.Scan()
			require.Equal(t, tcase.id, id, "Scan(%q) = (%s), want (%s)", tcase.in, tokenName(id), tokenName(tcase.id))
			require.Equal(t, tcase.want, string(got))
		})
	}
}

func TestSplitStatement(t *testing.T) {
	testcases := []struct {
		in  string
		sql string
		rem string
	}{{
		in:  "select * from table",
		sql: "select * from table",
	}, {
		in:  "select * from table; ",
		sql: "select * from table",
		rem: " ",
	}, {
		in:  "select * from table; select * from table2;",
		sql: "select * from table",
		rem: " select * from table2;",
	}, {
		in:  "select * from /* comment */ table;",
		sql: "select * from /* comment */ table",
	}, {
		in:  "select * from /* comment ; */ table;",
		sql: "select * from /* comment ; */ table",
	}, {
		in:  "select * from table where semi = ';';",
		sql: "select * from table where semi = ';'",
	}, {
		in:  "-- select * from table",
		sql: "-- select * from table",
	}, {
		in:  " ",
		sql: " ",
	}, {
		in:  "",
		sql: "",
	}}

	for _, tcase := range testcases {
		t.Run(tcase.in, func(t *testing.T) {
			sql, rem, err := sql_parser.SplitStatement(tcase.in, dialect.MYSQL)
			if err != nil {
				t.Errorf("EndOfStatementPosition(%s): ERROR: %v", tcase.in, err)
				return
			}

			if tcase.sql != sql {
				t.Errorf("EndOfStatementPosition(%s) got sql \"%s\" want \"%s\"", tcase.in, sql, tcase.sql)
			}

			if tcase.rem != rem {
				t.Errorf("EndOfStatementPosition(%s) got remainder \"%s\" want \"%s\"", tcase.in, rem, tcase.rem)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	testcases := []struct {
		version string
		in      string
		id      []int
	}{{
		version: "50709",
		in:      "/*!80102 SELECT*/ FROM IN EXISTS",
		id:      []int{mysql.FROM, mysql.IN, mysql.EXISTS, 0},
	}, {
		version: "80101",
		in:      "/*!80102 SELECT*/ FROM IN EXISTS",
		id:      []int{mysql.FROM, mysql.IN, mysql.EXISTS, 0},
	}, {
		version: "80201",
		in:      "/*!80102 SELECT*/ FROM IN EXISTS",
		id:      []int{mysql.SELECT, mysql.FROM, mysql.IN, mysql.EXISTS, 0},
	}, {
		version: "80102",
		in:      "/*!80102 SELECT*/ FROM IN EXISTS",
		id:      []int{mysql.SELECT, mysql.FROM, mysql.IN, mysql.EXISTS, 0},
	}}

	for _, tcase := range testcases {
		t.Run(tcase.version+"_"+tcase.in, func(t *testing.T) {
			mysql.MySQLVersion = tcase.version
			tok, err := sql_parser.NewStringTokenizer(tcase.in, dialect.MYSQL)
			if err != nil {
				t.Fatalf("%q", err)
			}
			for _, expectedID := range tcase.id {
				id, _ := tok.Scan()
				require.Equal(t, expectedID, id)
			}
		})
	}
}

func TestExtractMySQLComment(t *testing.T) {
	testcases := []struct {
		comment string
		version string
	}{{
		comment: "/*!50108 SELECT * FROM */",
		version: "50108",
	}, {
		comment: "/*!5018 SELECT * FROM */",
		version: "",
	}, {
		comment: "/*!SELECT * FROM */",
		version: "",
	}}

	for _, tcase := range testcases {
		t.Run(tcase.version, func(t *testing.T) {
			output, _ := mysql.ExtractMysqlComment(tcase.comment)
			require.Equal(t, tcase.version, output)
		})
	}
}

func TestIntegerAndID(t *testing.T) {
	testcases := []struct {
		in  string
		id  int
		out string
	}{{
		in: "334",
		id: mysql.INTEGRAL,
	}, {
		in: "33.4",
		id: mysql.DECIMAL,
	}, {
		in: "0x33",
		id: mysql.HEXNUM,
	}, {
		in: "33e4",
		id: mysql.FLOAT,
	}, {
		in: "33.4e-3",
		id: mysql.FLOAT,
	}, {
		in: "33t4",
		id: mysql.ID,
	}, {
		in: "0x2et3",
		id: mysql.ID,
	}, {
		in:  "3e2t3",
		id:  mysql.LEX_ERROR,
		out: "3e2",
	}, {
		in:  "3.2t",
		id:  mysql.LEX_ERROR,
		out: "3.2",
	}}

	for _, tcase := range testcases {
		t.Run(tcase.in, func(t *testing.T) {
			tkn, err := sql_parser.NewStringTokenizer(tcase.in, dialect.MYSQL)
			if err != nil {
				t.Fatalf("%q", err)
			}
			id, out := tkn.Scan()
			require.Equal(t, tcase.id, id)
			expectedOut := tcase.out
			if expectedOut == "" {
				expectedOut = tcase.in
			}
			require.Equal(t, expectedOut, out)
		})
	}
}
