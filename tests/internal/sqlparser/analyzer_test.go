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

package sqlparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/usalko/sent/internal/sqlparser"
)

func TestPreview(t *testing.T) {
	testcases := []struct {
		sql  string
		want sqlparser.StatementType
	}{
		{"select ...", sqlparser.StmtSelect},
		{"    select ...", sqlparser.StmtSelect},
		{"(select ...", sqlparser.StmtSelect},
		{"( select ...", sqlparser.StmtSelect},
		{"insert ...", sqlparser.StmtInsert},
		{"replace ....", sqlparser.StmtReplace},
		{"   update ...", sqlparser.StmtUpdate},
		{"Update", sqlparser.StmtUpdate},
		{"UPDATE ...", sqlparser.StmtUpdate},
		{"\n\t    delete ...", sqlparser.StmtDelete},
		{"", sqlparser.StmtUnknown},
		{" ", sqlparser.StmtUnknown},
		{"begin", sqlparser.StmtBegin},
		{" begin", sqlparser.StmtBegin},
		{" begin ", sqlparser.StmtBegin},
		{"\n\t begin ", sqlparser.StmtBegin},
		{"... begin ", sqlparser.StmtUnknown},
		{"begin ...", sqlparser.StmtUnknown},
		{"begin /* ... */", sqlparser.StmtBegin},
		{"begin /* ... *//*test*/", sqlparser.StmtBegin},
		{"begin;", sqlparser.StmtBegin},
		{"begin ;", sqlparser.StmtBegin},
		{"begin; /*...*/", sqlparser.StmtBegin},
		{"start transaction", sqlparser.StmtBegin},
		{"commit", sqlparser.StmtCommit},
		{"commit /*...*/", sqlparser.StmtCommit},
		{"rollback", sqlparser.StmtRollback},
		{"rollback /*...*/", sqlparser.StmtRollback},
		{"create", sqlparser.StmtDDL},
		{"alter", sqlparser.StmtDDL},
		{"rename", sqlparser.StmtDDL},
		{"drop", sqlparser.StmtDDL},
		{"set", sqlparser.StmtSet},
		{"show", sqlparser.StmtShow},
		{"use", sqlparser.StmtUse},
		{"analyze", sqlparser.StmtOther},
		{"describe", sqlparser.StmtExplain},
		{"desc", sqlparser.StmtExplain},
		{"explain", sqlparser.StmtExplain},
		{"repair", sqlparser.StmtOther},
		{"optimize", sqlparser.StmtOther},
		{"grant", sqlparser.StmtPriv},
		{"revoke", sqlparser.StmtPriv},
		{"truncate", sqlparser.StmtDDL},
		{"flush", sqlparser.StmtFlush},
		{"unknown", sqlparser.StmtUnknown},

		{"/* leading comment */ select ...", sqlparser.StmtSelect},
		{"/* leading comment */ (select ...", sqlparser.StmtSelect},
		{"/* leading comment */ /* leading comment 2 */ select ...", sqlparser.StmtSelect},
		{"/*! MySQL-specific comment */", sqlparser.StmtComment},
		{"/*!50708 MySQL-version comment */", sqlparser.StmtComment},
		{"-- leading single line comment \n select ...", sqlparser.StmtSelect},
		{"-- leading single line comment \n -- leading single line comment 2\n select ...", sqlparser.StmtSelect},

		{"/* leading comment no end select ...", sqlparser.StmtUnknown},
		{"-- leading single line comment no end select ...", sqlparser.StmtUnknown},
		{"/*!40000 ALTER TABLE `t1` DISABLE KEYS */", sqlparser.StmtComment},
	}
	for _, tcase := range testcases {
		if got := sqlparser.Preview(tcase.sql); got != tcase.want {
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
		if got := sqlparser.IsDML(tcase.sql); got != tcase.want {
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
		stmt, err := sqlparser.Parse(tcase.sql)
		assert.NoError(t, err)
		var expr sqlparser.Expr
		if where := stmt.(*sqlparser.Select).Where; where != nil {
			expr = where.Expr
		}
		splits := sqlparser.SplitAndExpression(nil, expr)
		var got []string
		for _, split := range splits {
			got = append(got, sqlparser.String(split))
		}
		assert.Equal(t, tcase.out, got)
	}
}

func TestAndExpressions(t *testing.T) {
	greaterThanExpr := &sqlparser.ComparisonExpr{
		Operator: sqlparser.GreaterThanOp,
		Left: &sqlparser.ColName{
			Name: sqlparser.NewColIdent("val"),
			Qualifier: sqlparser.TableName{
				Name: sqlparser.NewTableIdent("a"),
			},
		},
		Right: &sqlparser.ColName{
			Name: sqlparser.NewColIdent("val"),
			Qualifier: sqlparser.TableName{
				Name: sqlparser.NewTableIdent("b"),
			},
		},
	}
	equalExpr := &sqlparser.ComparisonExpr{
		Operator: sqlparser.EqualOp,
		Left: &sqlparser.ColName{
			Name: sqlparser.NewColIdent("id"),
			Qualifier: sqlparser.TableName{
				Name: sqlparser.NewTableIdent("a"),
			},
		},
		Right: &sqlparser.ColName{
			Name: sqlparser.NewColIdent("id"),
			Qualifier: sqlparser.TableName{
				Name: sqlparser.NewTableIdent("b"),
			},
		},
	}
	testcases := []struct {
		name           string
		expressions    sqlparser.Exprs
		expectedOutput sqlparser.Expr
	}{
		{
			name:           "empty input",
			expressions:    nil,
			expectedOutput: nil,
		}, {
			name: "two equal inputs",
			expressions: sqlparser.Exprs{
				greaterThanExpr,
				equalExpr,
				equalExpr,
			},
			expectedOutput: &sqlparser.AndExpr{
				Left:  greaterThanExpr,
				Right: equalExpr,
			},
		},
		{
			name: "two equal inputs",
			expressions: sqlparser.Exprs{
				equalExpr,
				equalExpr,
			},
			expectedOutput: equalExpr,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			output := sqlparser.AndExpressions(testcase.expressions...)
			assert.Equal(t, sqlparser.String(testcase.expectedOutput), sqlparser.String(output))
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
		name, err := sqlparser.TableFromStatement(tc.in)
		var got string
		if err != nil {
			got = err.Error()
		} else {
			got = sqlparser.String(name)
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
		tree, err := sqlparser.Parse(tc.in)
		if err != nil {
			t.Error(err)
			continue
		}
		out := sqlparser.GetTableName(tree.(*sqlparser.Select).From[0].(*sqlparser.AliasedTableExpr).Expr)
		if out.String() != tc.out {
			t.Errorf("GetTableName('%s'): %s, want %s", tc.in, out, tc.out)
		}
	}
}

func TestIsColName(t *testing.T) {
	testcases := []struct {
		in  sqlparser.Expr
		out bool
	}{{
		in:  &sqlparser.ColName{},
		out: true,
	}, {
		in: sqlparser.NewHexLiteral(""),
	}}
	for _, tc := range testcases {
		out := sqlparser.IsColName(tc.in)
		if out != tc.out {
			t.Errorf("IsColName(%T): %v, want %v", tc.in, out, tc.out)
		}
	}
}

func TestIsNull(t *testing.T) {
	testcases := []struct {
		in  sqlparser.Expr
		out bool
	}{{
		in:  &sqlparser.NullVal{},
		out: true,
	}, {
		in: sqlparser.NewStrLiteral(""),
	}}
	for _, tc := range testcases {
		out := sqlparser.IsNull(tc.in)
		if out != tc.out {
			t.Errorf("IsNull(%T): %v, want %v", tc.in, out, tc.out)
		}
	}
}
