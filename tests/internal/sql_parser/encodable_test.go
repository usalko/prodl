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
	"strings"
	"testing"

	"github.com/usalko/prodl/internal/sql_parser"
	"github.com/usalko/prodl/internal/sql_parser/ast"
	"github.com/usalko/prodl/internal/sql_types"
)

func TestEncodable(t *testing.T) {
	tcases := []struct {
		in  sql_parser.Encodable
		out string
	}{{
		in: sql_parser.InsertValues{{
			sql_types.NewInt64(1),
			sql_types.NewVarBinary("foo('a')"),
		}, {
			sql_types.NewInt64(2),
			sql_types.NewVarBinary("bar(`b`)"),
		}},
		out: "(1, 'foo(\\'a\\')'), (2, 'bar(`b`)')",
	}, {
		// Single column.
		in: &sql_parser.TupleEqualityList{
			Columns: []ast.ColIdent{ast.NewColIdent("pk")},
			Rows: [][]sql_types.Value{
				{sql_types.NewInt64(1)},
				{sql_types.NewVarBinary("aa")},
			},
		},
		out: "pk in (1, 'aa')",
	}, {
		// Multiple columns.
		in: &sql_parser.TupleEqualityList{
			Columns: []ast.ColIdent{ast.NewColIdent("pk1"), ast.NewColIdent("pk2")},
			Rows: [][]sql_types.Value{
				{
					sql_types.NewInt64(1),
					sql_types.NewVarBinary("aa"),
				},
				{
					sql_types.NewInt64(2),
					sql_types.NewVarBinary("bb"),
				},
			},
		},
		out: "(pk1 = 1 and pk2 = 'aa') or (pk1 = 2 and pk2 = 'bb')",
	}}
	for _, tcase := range tcases {
		buf := new(strings.Builder)
		tcase.in.EncodeSQL(buf)
		if out := buf.String(); out != tcase.out {
			t.Errorf("EncodeSQL(%v): %s, want %s", tcase.in, out, tcase.out)
		}
	}
}
