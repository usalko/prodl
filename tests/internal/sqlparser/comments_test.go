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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/usalko/sent/internal/sql_parser"
)

func TestSplitComments(t *testing.T) {
	var testCases = []struct {
		input, outSQL, outLeadingComments, outTrailingComments string
	}{{
		input:               "/",
		outSQL:              "/",
		outLeadingComments:  "",
		outTrailingComments: "",
	}, {
		input:               "*/",
		outSQL:              "*/",
		outLeadingComments:  "",
		outTrailingComments: "",
	}, {
		input:               "/*/",
		outSQL:              "/*/",
		outLeadingComments:  "",
		outTrailingComments: "",
	}, {
		input:               "a*/",
		outSQL:              "a*/",
		outLeadingComments:  "",
		outTrailingComments: "",
	}, {
		input:               "*a*/",
		outSQL:              "*a*/",
		outLeadingComments:  "",
		outTrailingComments: "",
	}, {
		input:               "**a*/",
		outSQL:              "**a*/",
		outLeadingComments:  "",
		outTrailingComments: "",
	}, {
		input:               "/*b**a*/",
		outSQL:              "",
		outLeadingComments:  "",
		outTrailingComments: "/*b**a*/",
	}, {
		input:               "/*a*/",
		outSQL:              "",
		outLeadingComments:  "",
		outTrailingComments: "/*a*/",
	}, {
		input:               "/**/",
		outSQL:              "",
		outLeadingComments:  "",
		outTrailingComments: "/**/",
	}, {
		input:               "/*b*/ /*a*/",
		outSQL:              "",
		outLeadingComments:  "",
		outTrailingComments: "/*b*/ /*a*/",
	}, {
		input:               "/* before */ foo /* bar */",
		outSQL:              "foo",
		outLeadingComments:  "/* before */ ",
		outTrailingComments: " /* bar */",
	}, {
		input:               "/* before1 */ /* before2 */ foo /* after1 */ /* after2 */",
		outSQL:              "foo",
		outLeadingComments:  "/* before1 */ /* before2 */ ",
		outTrailingComments: " /* after1 */ /* after2 */",
	}, {
		input:               "/** before */ foo /** bar */",
		outSQL:              "foo",
		outLeadingComments:  "/** before */ ",
		outTrailingComments: " /** bar */",
	}, {
		input:               "/*** before */ foo /*** bar */",
		outSQL:              "foo",
		outLeadingComments:  "/*** before */ ",
		outTrailingComments: " /*** bar */",
	}, {
		input:               "/** before **/ foo /** bar **/",
		outSQL:              "foo",
		outLeadingComments:  "/** before **/ ",
		outTrailingComments: " /** bar **/",
	}, {
		input:               "/*** before ***/ foo /*** bar ***/",
		outSQL:              "foo",
		outLeadingComments:  "/*** before ***/ ",
		outTrailingComments: " /*** bar ***/",
	}, {
		input:               " /*** before ***/ foo /*** bar ***/ ",
		outSQL:              "foo",
		outLeadingComments:  "/*** before ***/ ",
		outTrailingComments: " /*** bar ***/",
	}, {
		input:               "*** bar ***/",
		outSQL:              "*** bar ***/",
		outLeadingComments:  "",
		outTrailingComments: "",
	}, {
		input:               " foo ",
		outSQL:              "foo",
		outLeadingComments:  "",
		outTrailingComments: "",
	}, {
		input:               "select 1 from t where col = '*//*'",
		outSQL:              "select 1 from t where col = '*//*'",
		outLeadingComments:  "",
		outTrailingComments: "",
	}, {
		input:               "/*! select 1 */",
		outSQL:              "/*! select 1 */",
		outLeadingComments:  "",
		outTrailingComments: "",
	}}
	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			gotSQL, gotComments := sql_parser.SplitMarginComments(testCase.input)
			gotLeadingComments, gotTrailingComments := gotComments.Leading, gotComments.Trailing

			if gotSQL != testCase.outSQL {
				t.Errorf("test input: '%s', got SQL\n%+v, want\n%+v", testCase.input, gotSQL, testCase.outSQL)
			}
			if gotLeadingComments != testCase.outLeadingComments {
				t.Errorf("test input: '%s', got LeadingComments\n%+v, want\n%+v", testCase.input, gotLeadingComments, testCase.outLeadingComments)
			}
			if gotTrailingComments != testCase.outTrailingComments {
				t.Errorf("test input: '%s', got TrailingComments\n%+v, want\n%+v", testCase.input, gotTrailingComments, testCase.outTrailingComments)
			}
		})
	}
}

func TestStripLeadingComments(t *testing.T) {
	var testCases = []struct {
		input, outSQL string
	}{{
		input:  "/",
		outSQL: "/",
	}, {
		input:  "*/",
		outSQL: "*/",
	}, {
		input:  "/*/",
		outSQL: "/*/",
	}, {
		input:  "/*a",
		outSQL: "/*a",
	}, {
		input:  "/*a*",
		outSQL: "/*a*",
	}, {
		input:  "/*a**",
		outSQL: "/*a**",
	}, {
		input:  "/*b**a*/",
		outSQL: "",
	}, {
		input:  "/*a*/",
		outSQL: "",
	}, {
		input:  "/**/",
		outSQL: "",
	}, {
		input:  "/*!*/",
		outSQL: "/*!*/",
	}, {
		input:  "/*!a*/",
		outSQL: "/*!a*/",
	}, {
		input:  "/*b*/ /*a*/",
		outSQL: "",
	}, {
		input: `/*b*/ --foo
bar`,
		outSQL: "bar",
	}, {
		input:  "foo /* bar */",
		outSQL: "foo /* bar */",
	}, {
		input:  "/* foo */ bar",
		outSQL: "bar",
	}, {
		input:  "-- /* foo */ bar",
		outSQL: "",
	}, {
		input:  "foo -- bar */",
		outSQL: "foo -- bar */",
	}, {
		input: `/*
foo */ bar`,
		outSQL: "bar",
	}, {
		input: `-- foo bar
a`,
		outSQL: "a",
	}, {
		input:  `-- foo bar`,
		outSQL: "",
	}}
	for _, testCase := range testCases {
		gotSQL := sql_parser.StripLeadingComments(testCase.input)

		if gotSQL != testCase.outSQL {
			t.Errorf("test input: '%s', got SQL\n%+v, want\n%+v", testCase.input, gotSQL, testCase.outSQL)
		}
	}
}

func TestExtractMysqlComment(t *testing.T) {
	var testCases = []struct {
		input, outSQL, outVersion string
	}{{
		input:      "/*!50708SET max_execution_time=5000 */",
		outSQL:     "SET max_execution_time=5000",
		outVersion: "50708",
	}, {
		input:      "/*!50708 SET max_execution_time=5000*/",
		outSQL:     "SET max_execution_time=5000",
		outVersion: "50708",
	}, {
		input:      "/*!50708* from*/",
		outSQL:     "* from",
		outVersion: "50708",
	}, {
		input:      "/*! SET max_execution_time=5000*/",
		outSQL:     "SET max_execution_time=5000",
		outVersion: "",
	}}
	for _, testCase := range testCases {
		gotVersion, gotSQL := sql_parser.ExtractMysqlComment(testCase.input)

		if gotVersion != testCase.outVersion {
			t.Errorf("test input: '%s', got version\n%+v, want\n%+v", testCase.input, gotVersion, testCase.outVersion)
		}
		if gotSQL != testCase.outSQL {
			t.Errorf("test input: '%s', got SQL\n%+v, want\n%+v", testCase.input, gotSQL, testCase.outSQL)
		}
	}
}

func TestExtractCommentDirectives(t *testing.T) {
	var testCases = []struct {
		input string
		vals  sql_parser.CommentDirectives
	}{{
		input: "",
		vals:  nil,
	}, {
		input: "/* not a vt comment */",
		vals:  sql_parser.CommentDirectives{},
	}, {
		input: "/*vt+ */",
		vals:  sql_parser.CommentDirectives{},
	}, {
		input: "/*vt+ SINGLE_OPTION */",
		vals: sql_parser.CommentDirectives{
			"SINGLE_OPTION": "true",
		},
	}, {
		input: "/*vt+ ONE_OPT TWO_OPT */",
		vals: sql_parser.CommentDirectives{
			"ONE_OPT": "true",
			"TWO_OPT": "true",
		},
	}, {
		input: "/*vt+ ONE_OPT */ /* other comment */ /*vt+ TWO_OPT */",
		vals: sql_parser.CommentDirectives{
			"ONE_OPT": "true",
			"TWO_OPT": "true",
		},
	}, {
		input: "/*vt+ ONE_OPT=abc TWO_OPT=def */",
		vals: sql_parser.CommentDirectives{
			"ONE_OPT": "abc",
			"TWO_OPT": "def",
		},
	}, {
		input: "/*vt+ ONE_OPT=true TWO_OPT=false */",
		vals: sql_parser.CommentDirectives{
			"ONE_OPT": "true",
			"TWO_OPT": "false",
		},
	}, {
		input: "/*vt+ ONE_OPT=true TWO_OPT=\"false\" */",
		vals: sql_parser.CommentDirectives{
			"ONE_OPT": "true",
			"TWO_OPT": "\"false\"",
		},
	}, {
		input: "/*vt+ RANGE_OPT=[a:b] ANOTHER ANOTHER_WITH_VALEQ=val= AND_ONE_WITH_EQ== */",
		vals: sql_parser.CommentDirectives{
			"RANGE_OPT":          "[a:b]",
			"ANOTHER":            "true",
			"ANOTHER_WITH_VALEQ": "val=",
			"AND_ONE_WITH_EQ":    "=",
		},
	}}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			sqls := []string{
				"select " + testCase.input + " 1 from dual",
				"update " + testCase.input + " t set i=i+1",
				"delete " + testCase.input + " from t where id>1",
				"drop " + testCase.input + " table t",
				"create " + testCase.input + " table if not exists t (id int primary key)",
				"alter " + testCase.input + " table t add column c int not null",
				"create " + testCase.input + " view v as select * from t",
				"create " + testCase.input + " or replace view v as select * from t",
				"alter " + testCase.input + " view v as select * from t",
				"drop " + testCase.input + " view v",
			}
			for _, sql := range sqls {
				t.Run(sql, func(t *testing.T) {
					var comments *sql_parser.ParsedComments
					stmt, _ := sql_parser.Parse(sql)
					switch s := stmt.(type) {
					case *sql_parser.Select:
						comments = s.Comments
					case *sql_parser.Update:
						comments = s.Comments
					case *sql_parser.Delete:
						comments = s.Comments
					case *sql_parser.DropTable:
						comments = s.Comments
					case *sql_parser.AlterTable:
						comments = s.Comments
					case *sql_parser.CreateTable:
						comments = s.Comments
					case *sql_parser.CreateView:
						comments = s.Comments
					case *sql_parser.AlterView:
						comments = s.Comments
					case *sql_parser.DropView:
						comments = s.Comments
					default:
						t.Errorf("Unexpected statement type %+v", s)
					}

					vals := comments.Directives()
					if !reflect.DeepEqual(vals, testCase.vals) {
						t.Errorf("test input: '%v', got vals %T:\n%+v, want %T\n%+v", testCase.input, vals, vals, testCase.vals, testCase.vals)
					}
				})
			}
		})
	}

	d := sql_parser.CommentDirectives{
		"ONE_OPT": "true",
		"TWO_OPT": "false",
		"three":   "1",
		"four":    "2",
		"five":    "0",
		"six":     "true",
	}

	if !d.IsSet("ONE_OPT") {
		t.Errorf("d.IsSet(ONE_OPT) should be true")
	}

	if d.IsSet("TWO_OPT") {
		t.Errorf("d.IsSet(TWO_OPT) should be false")
	}

	if !d.IsSet("three") {
		t.Errorf("d.IsSet(three) should be true")
	}

	if d.IsSet("four") {
		t.Errorf("d.IsSet(four) should be false")
	}

	if d.IsSet("five") {
		t.Errorf("d.IsSet(five) should be false")
	}

	if !d.IsSet("six") {
		t.Errorf("d.IsSet(six) should be false")
	}
}

func TestSkipQueryPlanCacheDirective(t *testing.T) {
	stmt, _ := sql_parser.Parse("insert /*vt+ SKIP_QUERY_PLAN_CACHE=1 */ into user(id) values (1), (2)")
	if !sql_parser.SkipQueryPlanCacheDirective(stmt) {
		t.Errorf("d.SkipQueryPlanCacheDirective(stmt) should be true")
	}

	stmt, _ = sql_parser.Parse("insert into user(id) values (1), (2)")
	if sql_parser.SkipQueryPlanCacheDirective(stmt) {
		t.Errorf("d.SkipQueryPlanCacheDirective(stmt) should be false")
	}

	stmt, _ = sql_parser.Parse("update /*vt+ SKIP_QUERY_PLAN_CACHE=1 */ users set name=1")
	if !sql_parser.SkipQueryPlanCacheDirective(stmt) {
		t.Errorf("d.SkipQueryPlanCacheDirective(stmt) should be true")
	}

	stmt, _ = sql_parser.Parse("select /*vt+ SKIP_QUERY_PLAN_CACHE=1 */ * from users")
	if !sql_parser.SkipQueryPlanCacheDirective(stmt) {
		t.Errorf("d.SkipQueryPlanCacheDirective(stmt) should be true")
	}

	stmt, _ = sql_parser.Parse("delete /*vt+ SKIP_QUERY_PLAN_CACHE=1 */ from users")
	if !sql_parser.SkipQueryPlanCacheDirective(stmt) {
		t.Errorf("d.SkipQueryPlanCacheDirective(stmt) should be true")
	}
}

func TestIgnoreMaxPayloadSizeDirective(t *testing.T) {
	testCases := []struct {
		query    string
		expected bool
	}{
		{"insert /*vt+ IGNORE_MAX_PAYLOAD_SIZE=1 */ into user(id) values (1), (2)", true},
		{"insert into user(id) values (1), (2)", false},
		{"update /*vt+ IGNORE_MAX_PAYLOAD_SIZE=1 */ users set name=1", true},
		{"update users set name=1", false},
		{"select /*vt+ IGNORE_MAX_PAYLOAD_SIZE=1 */ * from users", true},
		{"select * from users", false},
		{"delete /*vt+ IGNORE_MAX_PAYLOAD_SIZE=1 */ from users", true},
		{"delete from users", false},
		{"show /*vt+ IGNORE_MAX_PAYLOAD_SIZE=1 */ create table users", false},
		{"show create table users", false},
	}

	for _, test := range testCases {
		t.Run(test.query, func(t *testing.T) {
			stmt, _ := sql_parser.Parse(test.query)
			got := sql_parser.IgnoreMaxPayloadSizeDirective(stmt)
			assert.Equalf(t, test.expected, got, fmt.Sprintf("IgnoreMaxPayloadSizeDirective(stmt) returned %v but expected %v", got, test.expected))
		})
	}
}

func TestIgnoreMaxMaxMemoryRowsDirective(t *testing.T) {
	testCases := []struct {
		query    string
		expected bool
	}{
		{"insert /*vt+ IGNORE_MAX_MEMORY_ROWS=1 */ into user(id) values (1), (2)", true},
		{"insert into user(id) values (1), (2)", false},
		{"update /*vt+ IGNORE_MAX_MEMORY_ROWS=1 */ users set name=1", true},
		{"update users set name=1", false},
		{"select /*vt+ IGNORE_MAX_MEMORY_ROWS=1 */ * from users", true},
		{"select * from users", false},
		{"delete /*vt+ IGNORE_MAX_MEMORY_ROWS=1 */ from users", true},
		{"delete from users", false},
		{"show /*vt+ IGNORE_MAX_MEMORY_ROWS=1 */ create table users", false},
		{"show create table users", false},
	}

	for _, test := range testCases {
		t.Run(test.query, func(t *testing.T) {
			stmt, _ := sql_parser.Parse(test.query)
			got := sql_parser.IgnoreMaxMaxMemoryRowsDirective(stmt)
			assert.Equalf(t, test.expected, got, fmt.Sprintf("IgnoreMaxPayloadSizeDirective(stmt) returned %v but expected %v", got, test.expected))
		})
	}
}
