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
	"github.com/usalko/sent/internal/sql_parser"
)

func TestLiteralID(t *testing.T) {
	testcases := []struct {
		in  string
		id  int
		out string
	}{{
		in:  "`aa`",
		id:  sql_parser.ID,
		out: "aa",
	}, {
		in:  "```a```",
		id:  sql_parser.ID,
		out: "`a`",
	}, {
		in:  "`a``b`",
		id:  sql_parser.ID,
		out: "a`b",
	}, {
		in:  "`a``b`c",
		id:  sql_parser.ID,
		out: "a`b",
	}, {
		in:  "`a``b",
		id:  sql_parser.LEX_ERROR,
		out: "a`b",
	}, {
		in:  "`a``b``",
		id:  sql_parser.LEX_ERROR,
		out: "a`b`",
	}, {
		in:  "``",
		id:  sql_parser.LEX_ERROR,
		out: "",
	}, {
		in:  "@x",
		id:  sql_parser.AT_ID,
		out: "x",
	}, {
		in:  "@@x",
		id:  sql_parser.AT_AT_ID,
		out: "x",
	}, {
		in:  "@@`x y`",
		id:  sql_parser.AT_AT_ID,
		out: "x y",
	}, {
		in:  "@@`@x @y`",
		id:  sql_parser.AT_AT_ID,
		out: "@x @y",
	}}

	for _, tcase := range testcases {
		t.Run(tcase.in, func(t *testing.T) {
			tkn := sql_parser.NewStringTokenizer(tcase.in)
			id, out := tkn.Scan()
			require.Equal(t, tcase.id, id)
			require.Equal(t, tcase.out, string(out))
		})
	}
}

func tokenName(id int) string {
	if id == sql_parser.STRING {
		return "STRING"
	} else if id == sql_parser.LEX_ERROR {
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
		id:   sql_parser.STRING,
		want: "",
	}, {
		in:   "''''",
		id:   sql_parser.STRING,
		want: "'",
	}, {
		in:   "'hello'",
		id:   sql_parser.STRING,
		want: "hello",
	}, {
		in:   "'\\n'",
		id:   sql_parser.STRING,
		want: "\n",
	}, {
		in:   "'\\nhello\\n'",
		id:   sql_parser.STRING,
		want: "\nhello\n",
	}, {
		in:   "'a''b'",
		id:   sql_parser.STRING,
		want: "a'b",
	}, {
		in:   "'a\\'b'",
		id:   sql_parser.STRING,
		want: "a'b",
	}, {
		in:   "'\\'",
		id:   sql_parser.LEX_ERROR,
		want: "'",
	}, {
		in:   "'",
		id:   sql_parser.LEX_ERROR,
		want: "",
	}, {
		in:   "'hello\\'",
		id:   sql_parser.LEX_ERROR,
		want: "hello'",
	}, {
		in:   "'hello",
		id:   sql_parser.LEX_ERROR,
		want: "hello",
	}, {
		in:   "'hello\\",
		id:   sql_parser.LEX_ERROR,
		want: "hello",
	}}

	for _, tcase := range testcases {
		t.Run(tcase.in, func(t *testing.T) {
			id, got := sql_parser.NewStringTokenizer(tcase.in).Scan()
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
			sql, rem, err := sql_parser.SplitStatement(tcase.in)
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
		id:      []int{sql_parser.FROM, sql_parser.IN, sql_parser.EXISTS, 0},
	}, {
		version: "80101",
		in:      "/*!80102 SELECT*/ FROM IN EXISTS",
		id:      []int{sql_parser.FROM, sql_parser.IN, sql_parser.EXISTS, 0},
	}, {
		version: "80201",
		in:      "/*!80102 SELECT*/ FROM IN EXISTS",
		id:      []int{sql_parser.SELECT, sql_parser.FROM, sql_parser.IN, sql_parser.EXISTS, 0},
	}, {
		version: "80102",
		in:      "/*!80102 SELECT*/ FROM IN EXISTS",
		id:      []int{sql_parser.SELECT, sql_parser.FROM, sql_parser.IN, sql_parser.EXISTS, 0},
	}}

	for _, tcase := range testcases {
		t.Run(tcase.version+"_"+tcase.in, func(t *testing.T) {
			sql_parser.MySQLVersion = tcase.version
			tok := sql_parser.NewStringTokenizer(tcase.in)
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
			output, _ := sql_parser.ExtractMysqlComment(tcase.comment)
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
		id: sql_parser.INTEGRAL,
	}, {
		in: "33.4",
		id: sql_parser.DECIMAL,
	}, {
		in: "0x33",
		id: sql_parser.HEXNUM,
	}, {
		in: "33e4",
		id: sql_parser.FLOAT,
	}, {
		in: "33.4e-3",
		id: sql_parser.FLOAT,
	}, {
		in: "33t4",
		id: sql_parser.ID,
	}, {
		in: "0x2et3",
		id: sql_parser.ID,
	}, {
		in:  "3e2t3",
		id:  sql_parser.LEX_ERROR,
		out: "3e2",
	}, {
		in:  "3.2t",
		id:  sql_parser.LEX_ERROR,
		out: "3.2",
	}}

	for _, tcase := range testcases {
		t.Run(tcase.in, func(t *testing.T) {
			tkn := sql_parser.NewStringTokenizer(tcase.in)
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
