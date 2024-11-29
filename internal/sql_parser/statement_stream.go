package sql_parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/usalko/prodl/internal/sql_parser/ast"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
	"github.com/usalko/prodl/internal/sql_parser/tokenizer"
)

const (
	PAGE_SIZE = 512
)

var (
	ErrIncompleteStatement = errors.New("ErrIncompleteStatement")
)

type StatementProcessor func(statementText string, statement ast.Statement, parseError error)

// Process text and return tail not processed (incomplete sql sentence)
func processText(text string, sql_dialect dialect.SqlDialect, processor StatementProcessor) (string, error) {
	var rawSql string = text
	switch strings.IndexByte(text, ';') {
	case -1: // if there is no semicolon, return blob as a whole
		stmt, err := Parse(text, sql_dialect)
		if stmt == nil {
			return text, ErrIncompleteStatement
		}
		processor(text, stmt, err)
		return "", nil
	case len(text) - 1: // if there's a single semicolon and it's the last character, return blob without it
		rawSql = text[:len(text)-1]
		stmt, err := Parse(rawSql, sql_dialect)
		if stmt == nil {
			return text, ErrIncompleteStatement
		}
		processor(rawSql, stmt, err)
		return "", nil
	}

	_tokenizer, err := NewStringTokenizer(text, sql_dialect)
	if err != nil {
		processor(text, nil, err)
		return "", nil
	}

	tkn := 0
	stmtBegin := 0
	statementIsEmpty := true
	for {
		tkn, _ = _tokenizer.Scan()
		switch tkn {
		case ';':
			rawSql = text[stmtBegin : _tokenizer.GetPos()-1]
			if !statementIsEmpty {
				stmt, err := Parse(rawSql, sql_dialect)
				processor(rawSql, stmt, err)
				statementIsEmpty = true
			}
			stmtBegin = _tokenizer.GetPos()
		case 0, tokenizer.EofChar:
			return text[stmtBegin:], ErrIncompleteStatement
		default:
			statementIsEmpty = false
		}
	}
}

// StatementStream split input stream into statements and call processor for every statement
func StatementStream(blob io.Reader, sql_dialect dialect.SqlDialect, processor StatementProcessor) error {
	if blob == nil {
		return fmt.Errorf("blob undefined (nil)")
	}
	page := make([]byte, PAGE_SIZE)
	statementBuffer := bytes.Buffer{}
	for {
		n, err := blob.Read(page)
		if n < PAGE_SIZE || err == io.EOF {
			statementBuffer.Write(page[:n])
			processText(statementBuffer.String(), sql_dialect, processor)
			return nil
		}
		statementBuffer.Write(page)
		buffTail, err := processText(statementBuffer.String(), sql_dialect, processor)
		if err == ErrIncompleteStatement {
			statementBuffer.Reset()
			statementBuffer.WriteString(buffTail)
		}
	}
}
