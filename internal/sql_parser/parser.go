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
	"io"
	"strings"

	"github.com/usalko/sent/internal/sql_parser/ast"
	"github.com/usalko/sent/internal/sql_parser/dialect"
	"github.com/usalko/sent/internal/sql_parser/mysql"
	"github.com/usalko/sent/internal/sql_parser_errors"
)

// Instructions for creating new types: If a type
// needs to satisfy an interface, declare that function
// along with that interface. This will help users
// identify the list of types to which they can assert
// those interfaces.
// If the member of a type has a string with a predefined
// list of values, declare those values as const following
// the type.
// For interfaces that define dummy functions to consolidate
// a set of types, define the function as iTypeName.
// This will help avoid name collisions.

// Parse2 parses the SQL in full and returns a Statement, which
// is the AST representation of the query, and a set of BindVars, which are all the
// bind variables that were found in the original SQL query. If a DDL statement
// is partially parsed but still contains a syntax error, the
// error is ignored and the DDL is returned anyway.
func Parse2(sql string) (ast.Statement, dialect.BindVars, error) {
	tokenizer, err := NewStringTokenizer(sql, dialect.MYSQL)
	if err != nil {
		return nil, nil, err
	}

	if val, _ := parsePooled(tokenizer, dialect.MYSQL); val != 0 {
		if tokenizer.GetPartialDDL() != nil {
			if typ, val := tokenizer.Scan(); typ != 0 {
				return nil, nil, fmt.Errorf("extra characters encountered after end of DDL: '%s'", string(val))
			}
			switch x := tokenizer.GetPartialDDL().(type) {
			case ast.DBDDLStatement:
				x.SetFullyParsed(false)
			case ast.DDLStatement:
				x.SetFullyParsed(false)
			}
			tokenizer.SetParseTree(tokenizer.GetPartialDDL())
			return tokenizer.GetParseTree(), tokenizer.GetBindVars(), nil
		}
		return nil, nil, sql_parser_errors.New(sql_parser_errors.Code_INVALID_ARGUMENT, tokenizer.GetLastError().Error())
	}
	if tokenizer.GetParseTree() == nil {
		return nil, nil, ErrEmpty
	}
	return tokenizer.GetParseTree(), tokenizer.GetBindVars(), nil
}

// TableFromStatement returns the qualified table name for the query.
// This works only for select statements.
func TableFromStatement(sql string) (ast.TableName, error) {
	stmt, err := Parse(sql)
	if err != nil {
		return ast.TableName{}, err
	}
	sel, ok := stmt.(*ast.Select)
	if !ok {
		return ast.TableName{}, fmt.Errorf("unrecognized statement: %s", sql)
	}
	if len(sel.From) != 1 {
		return ast.TableName{}, fmt.Errorf("table expression is complex")
	}
	aliased, ok := sel.From[0].(*ast.AliasedTableExpr)
	if !ok {
		return ast.TableName{}, fmt.Errorf("table expression is complex")
	}
	tableName, ok := aliased.Expr.(ast.TableName)
	if !ok {
		return ast.TableName{}, fmt.Errorf("table expression is complex")
	}
	return tableName, nil
}

// ParseExpr parses an expression and transforms it to an AST
func ParseExpr(sql string) (ast.Expr, error) {
	stmt, err := Parse("select " + sql)
	if err != nil {
		return nil, err
	}
	aliasedExpr := stmt.(*ast.Select).SelectExprs[0].(*ast.AliasedExpr)
	return aliasedExpr.Expr, err
}

// Parse behaves like Parse2 but does not return a set of bind variables
func Parse(sql string) (ast.Statement, error) {
	stmt, _, err := Parse2(sql)
	return stmt, err
}

// ParseStrictDDL is the same as Parse except it errors on
// partially parsed DDL statements.
func ParseStrictDDL(sql string) (ast.Statement, error) {
	tokenizer, err := NewStringTokenizer(sql, dialect.MYSQL)
	if err != nil {
		return nil, err
	}
	if val, _ := parsePooled(tokenizer, dialect.MYSQL); val != 0 {
		return nil, tokenizer.GetLastError()
	}
	if tokenizer.GetParseTree() == nil {
		return nil, ErrEmpty
	}
	return tokenizer.GetParseTree(), nil
}

// ParseTokenizer is a raw interface to parse from the given tokenizer.
// This does not used pooled parsers, and should not be used in general.
func ParseTokenizer(tokenizer dialect.Tokenizer, sql_dialect dialect.SqlDialect) (int, error) {
	return parse(tokenizer, sql_dialect)
}

// ParseNext parses a single SQL statement from the tokenizer
// returning a Statement which is the AST representation of the query.
// The tokenizer will always read up to the end of the statement, allowing for
// the next call to ParseNext to parse any subsequent SQL statements. When
// there are no more statements to parse, a error of io.EOF is returned.
func ParseNext(tokenizer dialect.Tokenizer) (ast.Statement, error) {
	return parseNext(tokenizer, false)
}

// ParseNextStrictDDL is the same as ParseNext except it errors on
// partially parsed DDL statements.
func ParseNextStrictDDL(tokenizer dialect.Tokenizer) (ast.Statement, error) {
	return parseNext(tokenizer, true)
}

func parseNext(tokenizer dialect.Tokenizer, strict bool) (ast.Statement, error) {
	if tokenizer.Cur() == ';' {
		tokenizer.Skip(1)
		tokenizer.SkipBlank()
	}
	if tokenizer.Cur() == dialect.EofChar {
		return nil, io.EOF
	}

	tokenizer.Reset()
	tokenizer.SetMulti(true)
	if val, _ := parsePooled(tokenizer, dialect.MYSQL); val != 0 {
		if tokenizer.GetPartialDDL() != nil && !strict {
			tokenizer.SetParseTree(tokenizer.GetPartialDDL())
			return tokenizer.GetParseTree(), nil
		}
		return nil, tokenizer.GetLastError()
	}
	if tokenizer.GetParseTree() == nil {
		return ParseNext(tokenizer)
	}
	return tokenizer.GetParseTree(), nil
}

// ErrEmpty is a sentinel error returned when parsing empty statements.
var ErrEmpty = sql_parser_errors.NewErrorf(sql_parser_errors.Code_INVALID_ARGUMENT, sql_parser_errors.EmptyQuery, "query was empty")

// SplitStatement returns the first sql statement up to either a ; or EOF
// and the remainder from the given buffer
func SplitStatement(blob string) (string, string, error) {
	tokenizer, err := NewStringTokenizer(blob, dialect.MYSQL)
	if err != nil {
		return "", "", err
	}

	tkn := 0
	for {
		tkn, _ = tokenizer.Scan()
		if tkn == 0 || tkn == ';' || tkn == dialect.EofChar {
			break
		}
	}
	if tokenizer.GetLastError() != nil {
		return "", "", tokenizer.GetLastError()
	}
	if tkn == ';' {
		return blob[:tokenizer.GetPos()-1], blob[tokenizer.GetPos():], nil
	}
	return blob, "", nil
}

// SplitStatementToPieces split raw sql statement that may have multi sql pieces to sql pieces
// returns the sql pieces blob contains; or error if sql cannot be parsed
func SplitStatementToPieces(blob string) (pieces []string, err error) {
	// fast path: the vast majority of SQL statements do not have semicolons in them
	if blob == "" {
		return nil, nil
	}
	switch strings.IndexByte(blob, ';') {
	case -1: // if there is no semicolon, return blob as a whole
		return []string{blob}, nil
	case len(blob) - 1: // if there's a single semicolon and it's the last character, return blob without it
		return []string{blob[:len(blob)-1]}, nil
	}

	pieces = make([]string, 0, 16)
	tokenizer, err := NewStringTokenizer(blob, dialect.MYSQL)
	if err != nil {
		return nil, err
	}

	tkn := 0
	var stmt string
	stmtBegin := 0
	emptyStatement := true
loop:
	for {
		tkn, _ = tokenizer.Scan()
		switch tkn {
		case ';':
			stmt = blob[stmtBegin : tokenizer.GetPos()-1]
			if !emptyStatement {
				pieces = append(pieces, stmt)
				emptyStatement = true
			}
			stmtBegin = tokenizer.GetPos()
		case 0, dialect.EofChar:
			blobTail := tokenizer.GetPos() - 1
			if stmtBegin < blobTail {
				stmt = blob[stmtBegin : blobTail+1]
				if !emptyStatement {
					pieces = append(pieces, stmt)
				}
			}
			break loop
		default:
			emptyStatement = false
		}
	}

	err = tokenizer.GetLastError()
	return
}

// NewStringTokenizer creates a new Tokenizer for the
// sql string.
func NewStringTokenizer(sql string, sql_dialect dialect.SqlDialect) (dialect.Tokenizer, error) {
	switch sql_dialect {
	case dialect.MYSQL:
		return mysql.NewMysqlStringTokenizer(sql), nil
	}
	return nil, fmt.Errorf("Sorry string tokenizer not found for dialect %s", sql_dialect)
}

func parsePooled(tokenizer dialect.Tokenizer, sql_dialect dialect.SqlDialect) (int, error) {
	switch sql_dialect {
	case dialect.MYSQL:
		return mysql.ParsePooled(tokenizer), nil
	}
	return 0, fmt.Errorf("Not implemented for dialect %s", sql_dialect)
}

func parse(tokenizer dialect.Tokenizer, sql_dialect dialect.SqlDialect) (int, error) {
	switch sql_dialect {
	case dialect.MYSQL:
		return mysql.Parse(tokenizer), nil
	}
	return 0, fmt.Errorf("Not implemented for dialect %s", sql_dialect)
}
