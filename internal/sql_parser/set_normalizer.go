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

	"github.com/usalko/sent/internal/sql_parser/ast"
	"github.com/usalko/sent/internal/sql_parser_errors"
)

type SetNormalizer struct {
	err error
}

func (n *SetNormalizer) rewriteSetComingUp(cursor *ast.Cursor) bool {
	set, ok := cursor.Node().(*ast.Set)
	if ok {
		for i, expr := range set.Exprs {
			exp, err := n.NormalizeSetExpr(expr)
			if err != nil {
				n.err = err
				return false
			}
			set.Exprs[i] = exp
		}
	}
	return true
}

func (n *SetNormalizer) NormalizeSetExpr(in *ast.SetExpr) (*ast.SetExpr, error) {
	switch in.Name.At { // using switch so we can use break
	case ast.DoubleAt:
		if in.Scope != ast.ImplicitScope {
			return nil, sql_parser_errors.Errorf(sql_parser_errors.Code_INVALID_ARGUMENT, "cannot use scope and @@")
		}
		switch {
		case strings.HasPrefix(in.Name.Lowered(), "session."):
			in.Name = createColumn(in.Name.Lowered()[8:])
			in.Scope = ast.SessionScope
		case strings.HasPrefix(in.Name.Lowered(), "global."):
			in.Name = createColumn(in.Name.Lowered()[7:])
			in.Scope = ast.GlobalScope
		case strings.HasPrefix(in.Name.Lowered(), "vitess_metadata."):
			in.Name = createColumn(in.Name.Lowered()[16:])
			in.Scope = ast.VitessMetadataScope
		default:
			in.Name.At = ast.NoAt
			in.Scope = ast.SessionScope
		}
		return in, nil
	case ast.SingleAt:
		if in.Scope != ast.ImplicitScope {
			return nil, sql_parser_errors.Errorf(sql_parser_errors.Code_INVALID_ARGUMENT, "cannot mix scope and user defined variables")
		}
		return in, nil
	case ast.NoAt:
		switch in.Scope {
		case ast.ImplicitScope:
			in.Scope = ast.SessionScope
		case ast.LocalScope:
			in.Scope = ast.SessionScope
		}
		return in, nil
	}
	panic("this should never happen")
}

func createColumn(str string) ast.ColIdent {
	size := len(str)
	if str[0] == '`' && str[size-1] == '`' {
		str = str[1 : size-1]
	}
	return ast.NewColIdent(str)
}
