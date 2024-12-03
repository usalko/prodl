package tests

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"github.com/usalko/prodl/internal/archive_stream"
	"github.com/usalko/prodl/internal/sql_parser"
	"github.com/usalko/prodl/internal/sql_parser/ast"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
)

type TextAndError struct {
	Text string
	Err  error
}

func check(err error, msgs ...any) {
	if err != nil {
		if len(msgs) == 0 {
			panic(err)
		} else if len(msgs) == 1 {
			panic(fmt.Errorf("%s: %s", msgs[0], err))
		} else {
			panic(fmt.Errorf("%s: %s", fmt.Sprintf(msgs[0].(string), msgs[1:]...), err))
		}
	}
}

func TestNopProcessing(t *testing.T) {
	fileName := "test_data/test_case1.txt.gz"
	respBody, err := os.Open(fileName)
	check(err, "File %s open error", fileName)

	reader := archive_stream.NewReader(respBody)

	for {
		entry, err := reader.GetNextEntry()
		if err == io.EOF {
			break
		}
		check(err, "unable to get next entry")

		if !entry.IsDir() {
			rc, err := entry.Open()
			defer func() {
				if err := rc.Close(); err != nil {
					log.Fatalf("close gzip entry reader fail: %s", err)
				}
			}()

			if err != nil {
				log.Fatalf("unable to open gzip file: %s", err)
			}
			statements := make([]ast.Statement, 0)
			parseErrors := make([]TextAndError, 0)
			sql_parser.StatementStream(rc, dialect.PSQL,
				func(statementText string, statement ast.Statement, parseError error) {
					if statement != nil {
						statements = append(statements, statement)
					}
					if parseError != nil {
						parseErrors = append(parseErrors, TextAndError{statementText, parseError})
					}
				})

			expectedStatementsCount := 15
			if len(statements) != expectedStatementsCount {
				t.Errorf("count of parsed statements is %v but expected %v", len(statements), expectedStatementsCount)
			}

			expectedErrorsCount := 0
			if len(parseErrors) > expectedErrorsCount {
				t.Errorf("nop processing has parse errors %v, %v", len(parseErrors), parseErrors)
			}
		}
	}
}
