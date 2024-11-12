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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/usalko/sent/internal/sql_parser"
)

func TestPreview(t *testing.T) {
	testcases := []struct {
		sql  string
		want sql_parser.StatementType
	}{
		{"select ...", sql_parser.StmtSelect},
		{"    select ...", sql_parser.StmtSelect},
		{"(select ...", sql_parser.StmtSelect},
		{"( select ...", sql_parser.StmtSelect},
		{"insert ...", sql_parser.StmtInsert},
		{"replace ....", sql_parser.StmtReplace},
		{"   update ...", sql_parser.StmtUpdate},
		{"Update", sql_parser.StmtUpdate},
		{"UPDATE ...", sql_parser.StmtUpdate},
		{"\n\t    delete ...", sql_parser.StmtDelete},
		{"", sql_parser.StmtUnknown},
		{" ", sql_parser.StmtUnknown},
		{"begin", sql_parser.StmtBegin},
		{" begin", sql_parser.StmtBegin},
		{" begin ", sql_parser.StmtBegin},
		{"\n\t begin ", sql_parser.StmtBegin},
		{"... begin ", sql_parser.StmtUnknown},
		{"begin ...", sql_parser.StmtUnknown},
		{"begin /* ... */", sql_parser.StmtBegin},
		{"begin /* ... *//*test*/", sql_parser.StmtBegin},
		{"begin;", sql_parser.StmtBegin},
		{"begin ;", sql_parser.StmtBegin},
		{"begin; /*...*/", sql_parser.StmtBegin},
		{"start transaction", sql_parser.StmtBegin},
		{"commit", sql_parser.StmtCommit},
		{"commit /*...*/", sql_parser.StmtCommit},
		{"rollback", sql_parser.StmtRollback},
		{"rollback /*...*/", sql_parser.StmtRollback},
		{"create", sql_parser.StmtDDL},
		{"alter", sql_parser.StmtDDL},
		{"rename", sql_parser.StmtDDL},
		{"drop", sql_parser.StmtDDL},
		{"set", sql_parser.StmtSet},
		{"show", sql_parser.StmtShow},
		{"use", sql_parser.StmtUse},
		{"analyze", sql_parser.StmtOther},
		{"describe", sql_parser.StmtExplain},
		{"desc", sql_parser.StmtExplain},
		{"explain", sql_parser.StmtExplain},
		{"repair", sql_parser.StmtOther},
		{"optimize", sql_parser.StmtOther},
		{"grant", sql_parser.StmtPriv},
		{"revoke", sql_parser.StmtPriv},
		{"truncate", sql_parser.StmtDDL},
		{"flush", sql_parser.StmtFlush},
		{"unknown", sql_parser.StmtUnknown},

		{"/* leading comment */ select ...", sql_parser.StmtSelect},
		{"/* leading comment */ (select ...", sql_parser.StmtSelect},
		{"/* leading comment */ /* leading comment 2 */ select ...", sql_parser.StmtSelect},
		{"/*! MySQL-specific comment */", sql_parser.StmtComment},
		{"/*!50708 MySQL-version comment */", sql_parser.StmtComment},
		{"-- leading single line comment \n select ...", sql_parser.StmtSelect},
		{"-- leading single line comment \n -- leading single line comment 2\n select ...", sql_parser.StmtSelect},

		{"/* leading comment no end select ...", sql_parser.StmtUnknown},
		{"-- leading single line comment no end select ...", sql_parser.StmtUnknown},
		{"/*!40000 ALTER TABLE `t1` DISABLE KEYS */", sql_parser.StmtComment},
	}
	for _, tcase := range testcases {
		if got := sql_parser.Preview(tcase.sql); got != tcase.want {
			t.Errorf("Preview(%s): %v, want %v", tcase.sql, got, tcase.want)
		}
	}
}

func TestIsDML(t *testing.T) {
	testcases := []struct {
		sql  string
		want bool
	}{
		{"   update ...", true},
		{"Update", true},
		{"UPDATE ...", true},
		{"\n\t    delete ...", true},
		{"insert ...", true},
		{"replace ...", true},
		{"select ...", false},
		{"    select ...", false},
		{"", false},
		{" ", false},
	}
	for _, tcase := range testcases {
		if got := sql_parser.IsDML(tcase.sql); got != tcase.want {
			t.Errorf("IsDML(%s): %v, want %v", tcase.sql, got, tcase.want)
		}
	}
}

func TestSplitAndExpression(t *testing.T) {
	testcases := []struct {
		sql string
		out []string
	}{{
		sql: "select * from t",
		out: nil,
	}, {
		sql: "select * from t where a = 1",
		out: []string{"a = 1"},
	}, {
		sql: "select * from t where a = 1 and b = 1",
		out: []string{"a = 1", "b = 1"},
	}, {
		sql: "select * from t where a = 1 and (b = 1 and c = 1)",
		out: []string{"a = 1", "b = 1", "c = 1"},
	}, {
		sql: "select * from t where a = 1 and (b = 1 or c = 1)",
		out: []string{"a = 1", "b = 1 or c = 1"},
	}, {
		sql: "select * from t where a = 1 and b = 1 or c = 1",
		out: []string{"a = 1 and b = 1 or c = 1"},
	}, {
		sql: "select * from t where a = 1 and b = 1 + (c = 1)",
		out: []string{"a = 1", "b = 1 + (c = 1)"},
	}, {
		sql: "select * from t where (a = 1 and ((b = 1 and c = 1)))",
		out: []string{"a = 1", "b = 1", "c = 1"},
	}}
	for _, tcase := range testcases {
		stmt, err := sql_parser.Parse(tcase.sql)
		assert.NoError(t, err)
		var expr sql_parser.Expr
		if where := stmt.(*sql_parser.Select).Where; where != nil {
			expr = where.Expr
		}
		splits := sql_parser.SplitAndExpression(nil, expr)
		var got []string
		for _, split := range splits {
			got = append(got, sql_parser.String(split))
		}
		assert.Equal(t, tcase.out, got)
	}
}

func TestAndExpressions(t *testing.T) {
	greaterThanExpr := &sql_parser.ComparisonExpr{
		Operator: sql_parser.GreaterThanOp,
		Left: &sql_parser.ColName{
			Name: sql_parser.NewColIdent("val"),
			Qualifier: sql_parser.TableName{
				Name: sql_parser.NewTableIdent("a"),
			},
		},
		Right: &sql_parser.ColName{
			Name: sql_parser.NewColIdent("val"),
			Qualifier: sql_parser.TableName{
				Name: sql_parser.NewTableIdent("b"),
			},
		},
	}
	equalExpr := &sql_parser.ComparisonExpr{
		Operator: sql_parser.EqualOp,
		Left: &sql_parser.ColName{
			Name: sql_parser.NewColIdent("id"),
			Qualifier: sql_parser.TableName{
				Name: sql_parser.NewTableIdent("a"),
			},
		},
		Right: &sql_parser.ColName{
			Name: sql_parser.NewColIdent("id"),
			Qualifier: sql_parser.TableName{
				Name: sql_parser.NewTableIdent("b"),
			},
		},
	}
	testcases := []struct {
		name           string
		expressions    sql_parser.Exprs
		expectedOutput sql_parser.Expr
	}{
		{
			name:           "empty input",
			expressions:    nil,
			expectedOutput: nil,
		}, {
			name: "two equal inputs",
			expressions: sql_parser.Exprs{
				greaterThanExpr,
				equalExpr,
				equalExpr,
			},
			expectedOutput: &sql_parser.AndExpr{
				Left:  greaterThanExpr,
				Right: equalExpr,
			},
		},
		{
			name: "two equal inputs",
			expressions: sql_parser.Exprs{
				equalExpr,
				equalExpr,
			},
			expectedOutput: equalExpr,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			output := sql_parser.AndExpressions(testcase.expressions...)
			assert.Equal(t, sql_parser.String(testcase.expectedOutput), sql_parser.String(output))
		})
	}
}

func TestTableFromStatement(t *testing.T) {
	testcases := []struct {
		in, out string
	}{{
		in:  "select * from t",
		out: "t",
	}, {
		in:  "select * from t.t",
		out: "t.t",
	}, {
		in:  "select * from t1, t2",
		out: "table expression is complex",
	}, {
		in:  "select * from (t)",
		out: "table expression is complex",
	}, {
		in:  "select * from t1 join t2",
		out: "table expression is complex",
	}, {
		in:  "select * from (select * from t) as tt",
		out: "table expression is complex",
	}, {
		in:  "update t set a=1",
		out: "unrecognized statement: update t set a=1",
	}, {
		in:  "bad query",
		out: "syntax error at position 4 near 'bad'",
	}}

	for _, tc := range testcases {
		name, err := sql_parser.TableFromStatement(tc.in)
		var got string
		if err != nil {
			got = err.Error()
		} else {
			got = sql_parser.String(name)
		}
		if got != tc.out {
			t.Errorf("TableFromStatement('%s'): %s, want %s", tc.in, got, tc.out)
		}
	}
}

func TestGetTableName(t *testing.T) {
	testcases := []struct {
		in, out string
	}{{
		in:  "select * from t",
		out: "t",
	}, {
		in:  "select * from t.t",
		out: "",
	}, {
		in:  "select * from (select * from t) as tt",
		out: "",
	}}

	for _, tc := range testcases {
		tree, err := sql_parser.Parse(tc.in)
		if err != nil {
			t.Error(err)
			continue
		}
		out := sql_parser.GetTableName(tree.(*sql_parser.Select).From[0].(*sql_parser.AliasedTableExpr).Expr)
		if out.String() != tc.out {
			t.Errorf("GetTableName('%s'): %s, want %s", tc.in, out, tc.out)
		}
	}
}

func TestIsColName(t *testing.T) {
	testcases := []struct {
		in  sql_parser.Expr
		out bool
	}{{
		in:  &sql_parser.ColName{},
		out: true,
	}, {
		in: sql_parser.NewHexLiteral(""),
	}}
	for _, tc := range testcases {
		out := sql_parser.IsColName(tc.in)
		if out != tc.out {
			t.Errorf("IsColName(%T): %v, want %v", tc.in, out, tc.out)
		}
	}
}

func TestIsNull(t *testing.T) {
	testcases := []struct {
		in  sql_parser.Expr
		out bool
	}{{
		in:  &sql_parser.NullVal{},
		out: true,
	}, {
		in: sql_parser.NewStrLiteral(""),
	}}
	for _, tc := range testcases {
		out := sql_parser.IsNull(tc.in)
		if out != tc.out {
			t.Errorf("IsNull(%T): %v, want %v", tc.in, out, tc.out)
		}
	}
}
