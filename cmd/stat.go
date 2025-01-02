package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/usalko/prodl/internal/archive_stream"
	"github.com/usalko/prodl/internal/sql_parser"
	"github.com/usalko/prodl/internal/sql_parser/ast"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
)

// statCmd represents the stat command
var statCmd = &cobra.Command{
	Use:   "stat",
	Short: "The 'stat' subcommand will stat request for dump.",
	Long: `The 'stat' subcommand stats a sql dump. For example:

'<cmd> stat dump-file-name.tar.gz'.`,
	Args: cobra.RangeArgs(1, MAX_COUNT_FOR_PROCESSING_FILES),
	Run: func(cmd *cobra.Command, args []string) {
		debugLevel, _ := cmd.Flags().GetInt("debug-level")
		// 1. Open file and detect dialect
		// 2. Request count of creating tables and they names

		sqlDialect := dialect.PSQL

		// Open reader and do StatementStream
		for _, fileName := range args {
			rootCmd.Printf("process file %v", fileName)
			stat, err := processFileForStat(fileName, sqlDialect, debugLevel)
			if err != nil {
				rootCmd.Println(" - fail")
				rootCmd.Println()
				rootCmd.PrintErrf("Error is %v", err)
				rootCmd.PrintErrln()
			} else {
				text := strings.Builder{}
				for k, _ := range stat.table_records {
					text.WriteString(k)
					text.WriteRune('\n')
				}
				rootCmd.Printf("%s", text.String())
			}
		}
	},
}

func init() {
	statCmd.Flags().IntP("debug-level", "d", 0, `
Debug level:

    0 no debug messages
	1 show debug messages
	2 show advanced debug messages

`)
	rootCmd.AddCommand(statCmd)
}

type DumpStat struct {
	table_records map[string]int
}

func processFileForStat(
	fileName string,
	sqlDialect dialect.SqlDialect,
	debugLevel int,
) (*DumpStat, error) {
	respBody, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("file %s open error (%v)", fileName, err)
	}
	defer respBody.Close()

	reader := archive_stream.NewReader(respBody)
	dumpStat := DumpStat{
		table_records: make(map[string]int, 100),
	}

	for {
		entry, err := reader.GetNextEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("unable to get next entry (%v)", err)
		}

		if !entry.IsDir() {
			rc, err := entry.Open()
			defer func() {
				if err := rc.Close(); err != nil {
					rootCmd.PrintErrf("close entry reader fail: %s", err)
				}
			}()

			if err != nil {
				return nil, fmt.Errorf("unable to open file: %s", err)
			}

			statementsCount := 0
			lastTime := time.Now()
			sql_parser.StatementStream(rc, sqlDialect,
				func(statementText string, statement ast.Statement, parseError error) {
					if parseError != nil {
						if debugLevel >= 1 {
							rootCmd.PrintErrf("parse sql statement:\n %s \n\nfail: %s\n", statementText, parseError)
						} else {
							rootCmd.PrintErrf("%s\n", parseError)
						}
					}
					createStatement, ok := statement.(*ast.CreateTable)
					if ok {
						dumpStat.table_records[createStatement.Table.Name.V] = 1
					}
					statementsCount++
					if debugLevel >= 2 {
						rootCmd.Printf("[%v] processed statements: %v\n", time.Since(lastTime), statementsCount)
					}
					lastTime = time.Now()
				})
		}
	}
	return &dumpStat, nil
}
