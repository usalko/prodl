package sql_parser

import (
	"errors"
	"fmt"
	"io"

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
func processText(_tokenizer tokenizer.Tokenizer, processor StatementProcessor) int {
	tkn := 0
	stmtBegin := 0
	statementIsEmpty := true
	for {
		tkn, _ = _tokenizer.Scan()
		switch tkn {
		case ';':
			if !statementIsEmpty {
				rawSql := _tokenizer.GetText(stmtBegin)
				stmt, err := Parse(rawSql, _tokenizer.GetDialect())
				processor(rawSql, stmt, err)
				statementIsEmpty = true
			}
			stmtBegin = _tokenizer.GetPos()
		case 0, tokenizer.EofChar:
			return stmtBegin
		default:
			statementIsEmpty = false
		}
	}
}

// StatementStream split input stream into statements and call processor for every statement
func StatementStream(blob io.Reader, sqlDialect dialect.SqlDialect, processor StatementProcessor) error {
	if blob == nil {
		return fmt.Errorf("blob undefined (nil)")
	}
	page := make([]byte, PAGE_SIZE)
	statementBuffer := tokenizer.BytesBuffer{}

	_tokenizer, err := NewBufferedTokenizer(&statementBuffer, sqlDialect)
	if err != nil {
		return fmt.Errorf("can't initialize a buffered tokenizer, error is %v", err)
	}

	for {
		n, err := blob.Read(page)
		if n < PAGE_SIZE || err == io.EOF {
			statementBuffer.Write(page[:n])
			processText(_tokenizer, processor)
			return nil
		}
		statementBuffer.Write(page)
		nextStmtPos := processText(_tokenizer, processor)
		// Reset do statementBuffer.ClipFrom(nextStmtPos)
		_tokenizer.ResetTo(nextStmtPos)
	}
}
