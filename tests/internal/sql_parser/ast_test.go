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
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	"github.com/usalko/prodl/internal/sql_parser"
	"github.com/usalko/prodl/internal/sql_parser/ast"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
)

func TestAppend(t *testing.T) {
	query := "select * from t where a = 1"
	tree, err := sql_parser.Parse(query, dialect.MYSQL)
	require.NoError(t, err)
	var b strings.Builder
	ast.Append(&b, tree)
	got := b.String()
	want := query
	if got != want {
		t.Errorf("Append: %s, want %s", got, want)
	}
	ast.Append(&b, tree)
	got = b.String()
	want = query + query
	if got != want {
		t.Errorf("Append: %s, want %s", got, want)
	}
}

func TestSelect(t *testing.T) {
	tree, err := sql_parser.Parse("select * from t where a = 1", dialect.MYSQL)
	require.NoError(t, err)
	expr := tree.(*ast.Select).Where.Expr

	sel := &ast.Select{}
	sel.AddWhere(expr)
	buf := ast.NewTrackedBuffer(nil)
	sel.Where.Format(buf)
	want := " where a = 1"
	if buf.String() != want {
		t.Errorf("where: %q, want %s", buf.String(), want)
	}
	sel.AddWhere(expr)
	buf = ast.NewTrackedBuffer(nil)
	sel.Where.Format(buf)
	want = " where a = 1"
	if buf.String() != want {
		t.Errorf("where: %q, want %s", buf.String(), want)
	}
	sel = &ast.Select{}
	sel.AddHaving(expr)
	buf = ast.NewTrackedBuffer(nil)
	sel.Having.Format(buf)
	want = " having a = 1"
	if buf.String() != want {
		t.Errorf("having: %q, want %s", buf.String(), want)
	}
	sel.AddHaving(expr)
	buf = ast.NewTrackedBuffer(nil)
	sel.Having.Format(buf)
	want = " having a = 1 and a = 1"
	if buf.String() != want {
		t.Errorf("having: %q, want %s", buf.String(), want)
	}

	tree, err = sql_parser.Parse("select * from t where a = 1 or b = 1", dialect.MYSQL)
	require.NoError(t, err)
	expr = tree.(*ast.Select).Where.Expr
	sel = &ast.Select{}
	sel.AddWhere(expr)
	buf = ast.NewTrackedBuffer(nil)
	sel.Where.Format(buf)
	want = " where a = 1 or b = 1"
	if buf.String() != want {
		t.Errorf("where: %q, want %s", buf.String(), want)
	}
	sel = &ast.Select{}
	sel.AddHaving(expr)
	buf = ast.NewTrackedBuffer(nil)
	sel.Having.Format(buf)
	want = " having a = 1 or b = 1"
	if buf.String() != want {
		t.Errorf("having: %q, want %s", buf.String(), want)
	}
}

func TestUpdate(t *testing.T) {
	tree, err := sql_parser.Parse("update t set a = 1", dialect.MYSQL)
	require.NoError(t, err)

	upd, ok := tree.(*ast.Update)
	require.True(t, ok)

	upd.AddWhere(&ast.ComparisonExpr{
		Left:     &ast.ColName{Name: ast.NewColIdent("b")},
		Operator: ast.EqualOp,
		Right:    ast.NewIntLiteral("2"),
	})
	assert.Equal(t, "update t set a = 1 where b = 2", ast.String(upd))

	upd.AddWhere(&ast.ComparisonExpr{
		Left:     &ast.ColName{Name: ast.NewColIdent("c")},
		Operator: ast.EqualOp,
		Right:    ast.NewIntLiteral("3"),
	})
	assert.Equal(t, "update t set a = 1 where b = 2 and c = 3", ast.String(upd))
}

func TestRemoveHints(t *testing.T) {
	for _, query := range []string{
		"select * from t use index (i)",
		"select * from t force index (i)",
	} {
		tree, err := sql_parser.Parse(query, dialect.MYSQL)
		if err != nil {
			t.Fatal(err)
		}
		sel := tree.(*ast.Select)
		sel.From = ast.TableExprs{
			sel.From[0].(*ast.AliasedTableExpr).RemoveHints(),
		}
		buf := ast.NewTrackedBuffer(nil)
		sel.Format(buf)
		if got, want := buf.String(), "select * from t"; got != want {
			t.Errorf("stripped query: %s, want %s", got, want)
		}
	}
}

func TestAddOrder(t *testing.T) {
	src, err := sql_parser.Parse("select foo, bar from baz order by foo", dialect.MYSQL)
	require.NoError(t, err)
	order := src.(*ast.Select).OrderBy[0]
	dst, err := sql_parser.Parse("select * from t", dialect.MYSQL)
	require.NoError(t, err)
	dst.(*ast.Select).AddOrder(order)
	buf := ast.NewTrackedBuffer(nil)
	dst.Format(buf)
	require.Equal(t, "select * from t order by foo asc", buf.String())
	dst, err = sql_parser.Parse("select * from t union select * from s", dialect.MYSQL)
	require.NoError(t, err)
	dst.(*ast.Union).AddOrder(order)
	buf = ast.NewTrackedBuffer(nil)
	dst.Format(buf)
	require.Equal(t, "select * from t union select * from s order by foo asc", buf.String())
}

func TestSetLimit(t *testing.T) {
	src, err := sql_parser.Parse("select foo, bar from baz limit 4", dialect.MYSQL)
	require.NoError(t, err)
	limit := src.(*ast.Select).Limit
	dst, err := sql_parser.Parse("select * from t", dialect.MYSQL)
	require.NoError(t, err)
	dst.(*ast.Select).SetLimit(limit)
	buf := ast.NewTrackedBuffer(nil)
	dst.Format(buf)
	require.Equal(t, "select * from t limit 4", buf.String())
	dst, err = sql_parser.Parse("select * from t union select * from s", dialect.MYSQL)
	require.NoError(t, err)
	dst.(*ast.Union).SetLimit(limit)
	buf = ast.NewTrackedBuffer(nil)
	dst.Format(buf)
	require.Equal(t, "select * from t union select * from s limit 4", buf.String())
}

func TestDDL(t *testing.T) {
	testcases := []struct {
		query    string
		output   ast.DDLStatement
		affected []string
	}{{
		query: "create table a",
		output: &ast.CreateTable{
			Table: ast.TableName{Name: ast.NewTableIdent("a")},
		},
		affected: []string{"a"},
	}, {
		query: "rename table a to b",
		output: &ast.RenameTable{
			TablePairs: []*ast.RenameTablePair{
				{
					FromTable: ast.TableName{Name: ast.NewTableIdent("a")},
					ToTable:   ast.TableName{Name: ast.NewTableIdent("b")},
				},
			},
		},
		affected: []string{"a", "b"},
	}, {
		query: "rename table a to b, c to d",
		output: &ast.RenameTable{
			TablePairs: []*ast.RenameTablePair{
				{
					FromTable: ast.TableName{Name: ast.NewTableIdent("a")},
					ToTable:   ast.TableName{Name: ast.NewTableIdent("b")},
				}, {
					FromTable: ast.TableName{Name: ast.NewTableIdent("c")},
					ToTable:   ast.TableName{Name: ast.NewTableIdent("d")},
				},
			},
		},
		affected: []string{"a", "b", "c", "d"},
	}, {
		query: "drop table a",
		output: &ast.DropTable{
			FromTables: ast.TableNames{
				ast.TableName{Name: ast.NewTableIdent("a")},
			},
		},
		affected: []string{"a"},
	}, {
		query: "drop table a, b",
		output: &ast.DropTable{
			FromTables: ast.TableNames{
				ast.TableName{Name: ast.NewTableIdent("a")},
				ast.TableName{Name: ast.NewTableIdent("b")},
			},
		},
		affected: []string{"a", "b"},
	}}
	for _, tcase := range testcases {
		got, err := sql_parser.Parse(tcase.query, dialect.MYSQL)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, tcase.output) {
			t.Errorf("%s: %v, want %v", tcase.query, got, tcase.output)
		}
		want := make(ast.TableNames, 0, len(tcase.affected))
		for _, t := range tcase.affected {
			want = append(want, ast.TableName{Name: ast.NewTableIdent(t)})
		}
		if affected := got.(ast.DDLStatement).AffectedTables(); !reflect.DeepEqual(affected, want) {
			t.Errorf("Affected(%s): %v, want %v", tcase.query, affected, want)
		}
	}
}

func TestSetAutocommitON(t *testing.T) {
	stmt, err := sql_parser.Parse("SET autocommit=ON", dialect.MYSQL)
	require.NoError(t, err)
	s, ok := stmt.(*ast.Set)
	if !ok {
		t.Errorf("SET statement is not Set: %T", s)
	}

	if len(s.Exprs) < 1 {
		t.Errorf("SET statement has no expressions")
	}

	e := s.Exprs[0]
	switch v := e.Expr.(type) {
	case *ast.Literal:
		if v.Type != ast.StrVal {
			t.Errorf("SET statement value is not StrVal: %T", v)
		}

		if "on" != v.Val {
			t.Errorf("SET statement value want: on, got: %s", v.Val)
		}
	default:
		t.Errorf("SET statement expression is not Literal: %T", e.Expr)
	}

	stmt, err = sql_parser.Parse("SET @@session.autocommit=ON", dialect.MYSQL)
	require.NoError(t, err)
	s, ok = stmt.(*ast.Set)
	if !ok {
		t.Errorf("SET statement is not Set: %T", s)
	}

	if len(s.Exprs) < 1 {
		t.Errorf("SET statement has no expressions")
	}

	e = s.Exprs[0]
	switch v := e.Expr.(type) {
	case *ast.Literal:
		if v.Type != ast.StrVal {
			t.Errorf("SET statement value is not StrVal: %T", v)
		}

		if "on" != v.Val {
			t.Errorf("SET statement value want: on, got: %s", v.Val)
		}
	default:
		t.Errorf("SET statement expression is not Literal: %T", e.Expr)
	}
}

func TestSetAutocommitOFF(t *testing.T) {
	stmt, err := sql_parser.Parse("SET autocommit=OFF", dialect.MYSQL)
	require.NoError(t, err)
	s, ok := stmt.(*ast.Set)
	if !ok {
		t.Errorf("SET statement is not Set: %T", s)
	}

	if len(s.Exprs) < 1 {
		t.Errorf("SET statement has no expressions")
	}

	e := s.Exprs[0]
	switch v := e.Expr.(type) {
	case *ast.Literal:
		if v.Type != ast.StrVal {
			t.Errorf("SET statement value is not StrVal: %T", v)
		}

		if "off" != v.Val {
			t.Errorf("SET statement value want: on, got: %s", v.Val)
		}
	default:
		t.Errorf("SET statement expression is not Literal: %T", e.Expr)
	}

	stmt, err = sql_parser.Parse("SET @@session.autocommit=OFF", dialect.MYSQL)
	require.NoError(t, err)
	s, ok = stmt.(*ast.Set)
	if !ok {
		t.Errorf("SET statement is not Set: %T", s)
	}

	if len(s.Exprs) < 1 {
		t.Errorf("SET statement has no expressions")
	}

	e = s.Exprs[0]
	switch v := e.Expr.(type) {
	case *ast.Literal:
		if v.Type != ast.StrVal {
			t.Errorf("SET statement value is not StrVal: %T", v)
		}

		if "off" != v.Val {
			t.Errorf("SET statement value want: on, got: %s", v.Val)
		}
	default:
		t.Errorf("SET statement expression is not Literal: %T", e.Expr)
	}

}

func TestWhere(t *testing.T) {
	var w *ast.Where
	buf := ast.NewTrackedBuffer(nil)
	w.Format(buf)
	if buf.String() != "" {
		t.Errorf("w.Format(nil): %q, want \"\"", buf.String())
	}
	w = ast.NewWhere(ast.WhereClause, nil)
	buf = ast.NewTrackedBuffer(nil)
	w.Format(buf)
	if buf.String() != "" {
		t.Errorf("w.Format(&Where{nil}: %q, want \"\"", buf.String())
	}
}

func TestIsAggregate(t *testing.T) {
	f := ast.FuncExpr{Name: ast.NewColIdent("avg")}
	if !f.IsAggregate() {
		t.Error("IsAggregate: false, want true")
	}

	f = ast.FuncExpr{Name: ast.NewColIdent("Avg")}
	if !f.IsAggregate() {
		t.Error("IsAggregate: false, want true")
	}

	f = ast.FuncExpr{Name: ast.NewColIdent("foo")}
	if f.IsAggregate() {
		t.Error("IsAggregate: true, want false")
	}
}

func TestIsImpossible(t *testing.T) {
	f := ast.ComparisonExpr{
		Operator: ast.NotEqualOp,
		Left:     ast.NewIntLiteral("1"),
		Right:    ast.NewIntLiteral("1"),
	}
	if !f.IsImpossible() {
		t.Error("IsImpossible: false, want true")
	}

	f = ast.ComparisonExpr{
		Operator: ast.EqualOp,
		Left:     ast.NewIntLiteral("1"),
		Right:    ast.NewIntLiteral("1"),
	}
	if f.IsImpossible() {
		t.Error("IsImpossible: true, want false")
	}

	f = ast.ComparisonExpr{
		Operator: ast.NotEqualOp,
		Left:     ast.NewIntLiteral("1"),
		Right:    ast.NewIntLiteral("2"),
	}
	if f.IsImpossible() {
		t.Error("IsImpossible: true, want false")
	}
}

func TestReplaceExpr(t *testing.T) {
	tcases := []struct {
		in, out string
	}{{
		in:  "select * from t where (select a from b)",
		out: ":a",
	}, {
		in:  "select * from t where (select a from b) and b",
		out: ":a and b",
	}, {
		in:  "select * from t where a and (select a from b)",
		out: "a and :a",
	}, {
		in:  "select * from t where (select a from b) or b",
		out: ":a or b",
	}, {
		in:  "select * from t where a or (select a from b)",
		out: "a or :a",
	}, {
		in:  "select * from t where not (select a from b)",
		out: "not :a",
	}, {
		in:  "select * from t where ((select a from b))",
		out: ":a",
	}, {
		in:  "select * from t where (select a from b) = 1",
		out: ":a = 1",
	}, {
		in:  "select * from t where a = (select a from b)",
		out: "a = :a",
	}, {
		in:  "select * from t where a like b escape (select a from b)",
		out: "a like b escape :a",
	}, {
		in:  "select * from t where (select a from b) between a and b",
		out: ":a between a and b",
	}, {
		in:  "select * from t where a between (select a from b) and b",
		out: "a between :a and b",
	}, {
		in:  "select * from t where a between b and (select a from b)",
		out: "a between b and :a",
	}, {
		in:  "select * from t where (select a from b) is null",
		out: ":a is null",
	}, {
		// exists should not replace.
		in:  "select * from t where exists (select a from b)",
		out: "exists (select a from b)",
	}, {
		in:  "select * from t where a in ((select a from b), 1)",
		out: "a in (:a, 1)",
	}, {
		in:  "select * from t where a in (0, (select a from b), 1)",
		out: "a in (0, :a, 1)",
	}, {
		in:  "select * from t where (select a from b) + 1",
		out: ":a + 1",
	}, {
		in:  "select * from t where 1+(select a from b)",
		out: "1 + :a",
	}, {
		in:  "select * from t where -(select a from b)",
		out: "-:a",
	}, {
		in:  "select * from t where interval (select a from b) aa",
		out: "interval :a aa",
	}, {
		in:  "select * from t where (select a from b) collate utf8",
		out: ":a collate utf8",
	}, {
		in:  "select * from t where func((select a from b), 1)",
		out: "func(:a, 1)",
	}, {
		in:  "select * from t where func(1, (select a from b), 1)",
		out: "func(1, :a, 1)",
	}, {
		in:  "select * from t where group_concat((select a from b), 1 order by a)",
		out: "group_concat(:a, 1 order by a asc)",
	}, {
		in:  "select * from t where group_concat(1 order by (select a from b), a)",
		out: "group_concat(1 order by :a asc, a asc)",
	}, {
		in:  "select * from t where group_concat(1 order by a, (select a from b))",
		out: "group_concat(1 order by a asc, :a asc)",
	}, {
		in:  "select * from t where substr(a, (select a from b), b)",
		out: "substr(a, :a, b)",
	}, {
		in:  "select * from t where substr(a, b, (select a from b))",
		out: "substr(a, b, :a)",
	}, {
		in:  "select * from t where convert((select a from b), json)",
		out: "convert(:a, json)",
	}, {
		in:  "select * from t where convert((select a from b) using utf8)",
		out: "convert(:a using utf8)",
	}, {
		in:  "select * from t where match((select a from b), 1) against (a)",
		out: "match(:a, 1) against (a)",
	}, {
		in:  "select * from t where match(1, (select a from b), 1) against (a)",
		out: "match(1, :a, 1) against (a)",
	}, {
		in:  "select * from t where match(1, a, 1) against ((select a from b))",
		out: "match(1, a, 1) against (:a)",
	}, {
		in:  "select * from t where case (select a from b) when a then b when b then c else d end",
		out: "case :a when a then b when b then c else d end",
	}, {
		in:  "select * from t where case a when (select a from b) then b when b then c else d end",
		out: "case a when :a then b when b then c else d end",
	}, {
		in:  "select * from t where case a when b then (select a from b) when b then c else d end",
		out: "case a when b then :a when b then c else d end",
	}, {
		in:  "select * from t where case a when b then c when (select a from b) then c else d end",
		out: "case a when b then c when :a then c else d end",
	}, {
		in:  "select * from t where case a when b then c when d then c else (select a from b) end",
		out: "case a when b then c when d then c else :a end",
	}}
	to := ast.NewArgument("a")
	for _, tcase := range tcases {
		tree, err := sql_parser.Parse(tcase.in, dialect.MYSQL)
		if err != nil {
			t.Fatal(err)
		}
		var from *ast.Subquery
		_ = ast.Walk(func(node ast.SQLNode) (kontinue bool, err error) {
			if sq, ok := node.(*ast.Subquery); ok {
				from = sq
				return false, nil
			}
			return true, nil
		}, tree)
		if from == nil {
			t.Fatalf("from is nil for %s", tcase.in)
		}
		expr := ast.ReplaceExpr(tree.(*ast.Select).Where.Expr, from, to)
		got := ast.String(expr)
		if tcase.out != got {
			t.Errorf("ReplaceExpr(%s): %s, want %s", tcase.in, got, tcase.out)
		}
	}
}

func TestColNameEqual(t *testing.T) {
	var c1, c2 *ast.ColName
	if c1.Equal(c2) {
		t.Error("nil columns equal, want unequal")
	}
	c1 = &ast.ColName{
		Name: ast.NewColIdent("aa"),
	}
	c2 = &ast.ColName{
		Name: ast.NewColIdent("bb"),
	}
	if c1.Equal(c2) {
		t.Error("columns equal, want unequal")
	}
	c2.Name = ast.NewColIdent("aa")
	if !c1.Equal(c2) {
		t.Error("columns unequal, want equal")
	}
}

func TestColIdent(t *testing.T) {
	str := ast.NewColIdent("Ab")
	if str.String() != "Ab" {
		t.Errorf("String=%s, want Ab", str.String())
	}
	if str.String() != "Ab" {
		t.Errorf("Val=%s, want Ab", str.String())
	}
	if str.Lowered() != "ab" {
		t.Errorf("Val=%s, want ab", str.Lowered())
	}
	if !str.Equal(ast.NewColIdent("aB")) {
		t.Error("str.Equal(NewColIdent(aB))=false, want true")
	}
	if !str.EqualString("ab") {
		t.Error("str.EqualString(ab)=false, want true")
	}
	str = ast.NewColIdent("")
	if str.Lowered() != "" {
		t.Errorf("Val=%s, want \"\"", str.Lowered())
	}
}

func TestColIdentMarshal(t *testing.T) {
	str := ast.NewColIdent("Ab")
	b, err := json.Marshal(str)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	want := `"Ab"`
	if got != want {
		t.Errorf("json.Marshal()= %s, want %s", got, want)
	}
	var out ast.ColIdent
	if err := json.Unmarshal(b, &out); err != nil {
		t.Errorf("Unmarshal err: %v, want nil", err)
	}
	if !reflect.DeepEqual(out, str) {
		t.Errorf("Unmarshal: %v, want %v", out, str)
	}
}

func TestColIdentSize(t *testing.T) {
	size := unsafe.Sizeof(ast.NewColIdent(""))
	want := 2*unsafe.Sizeof("") + 8
	if size != want {
		t.Errorf("Size of ColIdent: %d, want 32", want)
	}
}

func TestTableIdentMarshal(t *testing.T) {
	str := ast.NewTableIdent("Ab")
	b, err := json.Marshal(str)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	want := `"Ab"`
	if got != want {
		t.Errorf("json.Marshal()= %s, want %s", got, want)
	}
	var out ast.TableIdent
	if err := json.Unmarshal(b, &out); err != nil {
		t.Errorf("Unmarshal err: %v, want nil", err)
	}
	if !reflect.DeepEqual(out, str) {
		t.Errorf("Unmarshal: %v, want %v", out, str)
	}
}

func TestHexDecode(t *testing.T) {
	testcase := []struct {
		in, out string
	}{{
		in:  "313233",
		out: "123",
	}, {
		in:  "ag",
		out: "encoding/hex: invalid byte: U+0067 'g'",
	}, {
		in:  "777",
		out: "encoding/hex: odd length hex string",
	}}
	for _, tc := range testcase {
		out, err := ast.NewHexLiteral(tc.in).HexDecode()
		if err != nil {
			if err.Error() != tc.out {
				t.Errorf("Decode(%q): %v, want %s", tc.in, err, tc.out)
			}
			continue
		}
		if !bytes.Equal(out, []byte(tc.out)) {
			t.Errorf("Decode(%q): %s, want %s", tc.in, out, tc.out)
		}
	}
}

func TestCompliantName(t *testing.T) {
	testcases := []struct {
		in, out string
	}{{
		in:  "aa",
		out: "aa",
	}, {
		in:  "1a",
		out: "_a",
	}, {
		in:  "a1",
		out: "a1",
	}, {
		in:  "a.b",
		out: "a_b",
	}, {
		in:  ".ab",
		out: "_ab",
	}}
	for _, tc := range testcases {
		out := ast.NewColIdent(tc.in).CompliantName()
		if out != tc.out {
			t.Errorf("ColIdent(%s).CompliantNamt: %s, want %s", tc.in, out, tc.out)
		}
		out = ast.NewTableIdent(tc.in).CompliantName()
		if out != tc.out {
			t.Errorf("TableIdent(%s).CompliantNamt: %s, want %s", tc.in, out, tc.out)
		}
	}
}

func TestColumns_FindColumn(t *testing.T) {
	cols := ast.Columns{ast.NewColIdent("a"), ast.NewColIdent("c"), ast.NewColIdent("b"), ast.NewColIdent("0")}

	testcases := []struct {
		in  string
		out int
	}{{
		in:  "a",
		out: 0,
	}, {
		in:  "b",
		out: 2,
	},
		{
			in:  "0",
			out: 3,
		},
		{
			in:  "f",
			out: -1,
		}}

	for _, tc := range testcases {
		val := cols.FindColumn(ast.NewColIdent(tc.in))
		if val != tc.out {
			t.Errorf("FindColumn(%s): %d, want %d", tc.in, val, tc.out)
		}
	}
}

func TestSplitStatementToPieces(t *testing.T) {
	testcases := []struct {
		input  string
		output string
	}{{
		input:  "select * from table1; \t; \n; \n\t\t ;select * from table1;",
		output: "select * from table1;select * from table1",
	}, {
		input: "select * from table",
	}, {
		input:  "select * from table;",
		output: "select * from table",
	}, {
		input:  "select * from table;   ",
		output: "select * from table",
	}, {
		input:  "select * from table1; select * from table2;",
		output: "select * from table1; select * from table2",
	}, {
		input:  "select * from /* comment ; */ table;",
		output: "select * from /* comment ; */ table",
	}, {
		input:  "select * from table where semi = ';';",
		output: "select * from table where semi = ';'",
	}, {
		input:  "select * from table1;--comment;\nselect * from table2;",
		output: "select * from table1;--comment;\nselect * from table2",
	}, {
		input: "CREATE TABLE `total_data` (`id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'id', " +
			"`region` varchar(32) NOT NULL COMMENT 'region name, like zh; th; kepler'," +
			"`data_size` bigint NOT NULL DEFAULT '0' COMMENT 'data size;'," +
			"`createtime` datetime NOT NULL DEFAULT NOW() COMMENT 'create time;'," +
			"`comment` varchar(100) NOT NULL DEFAULT '' COMMENT 'comment'," +
			"PRIMARY KEY (`id`))",
	}}

	for _, tcase := range testcases {
		t.Run(tcase.input, func(t *testing.T) {
			if tcase.output == "" {
				tcase.output = tcase.input
			}

			stmtPieces, err := sql_parser.SplitStatementToPieces(tcase.input, dialect.MYSQL)
			require.NoError(t, err)

			out := strings.Join(stmtPieces, ";")
			require.Equal(t, tcase.output, out)
		})
	}
}

func TestTypeConversion(t *testing.T) {
	ct1 := &ast.ColumnType{Type: "BIGINT"}
	ct2 := &ast.ColumnType{Type: "bigint"}
	assert.Equal(t, ct1.SQLType(), ct2.SQLType())
}

func TestDefaultStatus(t *testing.T) {
	assert.Equal(t,
		ast.String(&ast.Default{ColName: "status"}),
		"default(`status`)")
}

func TestShowTableStatus(t *testing.T) {
	query := "Show Table Status FROM customer"
	tree, err := sql_parser.Parse(query, dialect.MYSQL)
	require.NoError(t, err)
	require.NotNil(t, tree)
}

func BenchmarkStringTraces(b *testing.B) {
	for _, trace := range []string{"django_queries.txt", "lobsters.sql.gz"} {
		b.Run(trace, func(b *testing.B) {
			queries := loadQueries(b, trace)
			if len(queries) > 10000 {
				queries = queries[:10000]
			}

			parsed := make([]ast.Statement, 0, len(queries))
			for _, q := range queries {
				pp, err := sql_parser.Parse(q, dialect.MYSQL)
				if err != nil {
					b.Fatal(err)
				}
				parsed = append(parsed, pp)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				for _, stmt := range parsed {
					_ = ast.String(stmt)
				}
			}
		})
	}
}
