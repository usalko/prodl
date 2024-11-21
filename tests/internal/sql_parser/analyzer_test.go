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

package sql_parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/usalko/sent/internal/sql_parser"
	"github.com/usalko/sent/internal/sql_parser/ast"
)

func TestPreview(t *testing.T) {
	testcases := []struct {
		sql  string
		want ast.StatementType
	}{
		{"select ...", ast.StmtSelect},
		{"    select ...", ast.StmtSelect},
		{"(select ...", ast.StmtSelect},
		{"( select ...", ast.StmtSelect},
		{"insert ...", ast.StmtInsert},
		{"replace ....", ast.StmtReplace},
		{"   update ...", ast.StmtUpdate},
		{"Update", ast.StmtUpdate},
		{"UPDATE ...", ast.StmtUpdate},
		{"\n\t    delete ...", ast.StmtDelete},
		{"", ast.StmtUnknown},
		{" ", ast.StmtUnknown},
		{"begin", ast.StmtBegin},
		{" begin", ast.StmtBegin},
		{" begin ", ast.StmtBegin},
		{"\n\t begin ", ast.StmtBegin},
		{"... begin ", ast.StmtUnknown},
		{"begin ...", ast.StmtUnknown},
		{"begin /* ... */", ast.StmtBegin},
		{"begin /* ... *//*test*/", ast.StmtBegin},
		{"begin;", ast.StmtBegin},
		{"begin ;", ast.StmtBegin},
		{"begin; /*...*/", ast.StmtBegin},
		{"start transaction", ast.StmtBegin},
		{"commit", ast.StmtCommit},
		{"commit /*...*/", ast.StmtCommit},
		{"rollback", ast.StmtRollback},
		{"rollback /*...*/", ast.StmtRollback},
		{"create", ast.StmtDDL},
		{"alter", ast.StmtDDL},
		{"rename", ast.StmtDDL},
		{"drop", ast.StmtDDL},
		{"set", ast.StmtSet},
		{"show", ast.StmtShow},
		{"use", ast.StmtUse},
		{"analyze", ast.StmtOther},
		{"describe", ast.StmtExplain},
		{"desc", ast.StmtExplain},
		{"explain", ast.StmtExplain},
		{"repair", ast.StmtOther},
		{"optimize", ast.StmtOther},
		{"grant", ast.StmtPriv},
		{"revoke", ast.StmtPriv},
		{"truncate", ast.StmtDDL},
		{"flush", ast.StmtFlush},
		{"unknown", ast.StmtUnknown},

		{"/* leading comment */ select ...", ast.StmtSelect},
		{"/* leading comment */ (select ...", ast.StmtSelect},
		{"/* leading comment */ /* leading comment 2 */ select ...", ast.StmtSelect},
		{"/*! MySQL-specific comment */", ast.StmtComment},
		{"/*!50708 MySQL-version comment */", ast.StmtComment},
		{"-- leading single line comment \n select ...", ast.StmtSelect},
		{"-- leading single line comment \n -- leading single line comment 2\n select ...", ast.StmtSelect},

		{"/* leading comment no end select ...", ast.StmtUnknown},
		{"-- leading single line comment no end select ...", ast.StmtUnknown},
		{"/*!40000 ALTER TABLE `t1` DISABLE KEYS */", ast.StmtComment},
	}
	for _, tcase := range testcases {
		if got := ast.Preview(tcase.sql); got != tcase.want {
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
		if got := ast.IsDML(tcase.sql); got != tcase.want {
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
		var expr ast.Expr
		if where := stmt.(*ast.Select).Where; where != nil {
			expr = where.Expr
		}
		splits := ast.SplitAndExpression(nil, expr)
		var got []string
		for _, split := range splits {
			got = append(got, ast.String(split))
		}
		assert.Equal(t, tcase.out, got)
	}
}

func TestAndExpressions(t *testing.T) {
	greaterThanExpr := &ast.ComparisonExpr{
		Operator: ast.GreaterThanOp,
		Left: &ast.ColName{
			Name: ast.NewColIdent("val"),
			Qualifier: ast.TableName{
				Name: ast.NewTableIdent("a"),
			},
		},
		Right: &ast.ColName{
			Name: ast.NewColIdent("val"),
			Qualifier: ast.TableName{
				Name: ast.NewTableIdent("b"),
			},
		},
	}
	equalExpr := &ast.ComparisonExpr{
		Operator: ast.EqualOp,
		Left: &ast.ColName{
			Name: ast.NewColIdent("id"),
			Qualifier: ast.TableName{
				Name: ast.NewTableIdent("a"),
			},
		},
		Right: &ast.ColName{
			Name: ast.NewColIdent("id"),
			Qualifier: ast.TableName{
				Name: ast.NewTableIdent("b"),
			},
		},
	}
	testcases := []struct {
		name           string
		expressions    ast.Exprs
		expectedOutput ast.Expr
	}{
		{
			name:           "empty input",
			expressions:    nil,
			expectedOutput: nil,
		}, {
			name: "two equal inputs",
			expressions: ast.Exprs{
				greaterThanExpr,
				equalExpr,
				equalExpr,
			},
			expectedOutput: &ast.AndExpr{
				Left:  greaterThanExpr,
				Right: equalExpr,
			},
		},
		{
			name: "two equal inputs",
			expressions: ast.Exprs{
				equalExpr,
				equalExpr,
			},
			expectedOutput: equalExpr,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			output := ast.AndExpressions(testcase.expressions...)
			assert.Equal(t, ast.String(testcase.expectedOutput), ast.String(output))
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
			got = ast.String(name)
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
		out := ast.GetTableName(tree.(*ast.Select).From[0].(*ast.AliasedTableExpr).Expr)
		if out.String() != tc.out {
			t.Errorf("GetTableName('%s'): %s, want %s", tc.in, out, tc.out)
		}
	}
}

func TestIsColName(t *testing.T) {
	testcases := []struct {
		in  ast.Expr
		out bool
	}{{
		in:  &ast.ColName{},
		out: true,
	}, {
		in: ast.NewHexLiteral(""),
	}}
	for _, tc := range testcases {
		out := ast.IsColName(tc.in)
		if out != tc.out {
			t.Errorf("IsColName(%T): %v, want %v", tc.in, out, tc.out)
		}
	}
}

func TestIsNull(t *testing.T) {
	testcases := []struct {
		in  ast.Expr
		out bool
	}{{
		in:  &ast.NullVal{},
		out: true,
	}, {
		in: ast.NewStrLiteral(""),
	}}
	for _, tc := range testcases {
		out := ast.IsNull(tc.in)
		if out != tc.out {
			t.Errorf("IsNull(%T): %v, want %v", tc.in, out, tc.out)
		}
	}
}
