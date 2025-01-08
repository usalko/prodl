package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/usalko/prodl/internal/archive_stream"
	"github.com/usalko/prodl/internal/sql_connection"
	"github.com/usalko/prodl/internal/sql_parser"
	"github.com/usalko/prodl/internal/sql_parser/ast"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
)

const MAX_COUNT_FOR_PROCESSING_FILES = 1024

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "The 'load' subcommand will load dump to the database.",
	Long: `The 'load' subcommand loads a sql dump to the database. For example:

'<cmd> load --to sqlite3://./local.sqlite3 dump-file-name.tar.gz'.`,
	Args: cobra.RangeArgs(1, MAX_COUNT_FOR_PROCESSING_FILES),
	Run: func(cmd *cobra.Command, args []string) {
		debugLevel, _ := cmd.Flags().GetInt("debug-level")
		targetSqlUrl, _ := cmd.Flags().GetString("target-sql-connection")
		sqlDialect, connectionOptions, err := (*dialect.SqlDialect).ParseUrl(nil, targetSqlUrl)
		if err != nil {
			rootCmd.PrintErrf("parse target url %v fail with error: %v\n", targetSqlUrl, err)
			return
		}
		connection, err := sql_connection.Connect(sqlDialect)
		if err != nil {
			rootCmd.PrintErrf("make connection structure for target url %v fail with error: %v\n", targetSqlUrl, err)
			return
		}

		err = connection.Establish(connectionOptions)
		if err != nil {
			rootCmd.PrintErrf("establish connection for target url %v fail with error: %v\n", targetSqlUrl, err)
			return
		}

		// Test connection to the database
		err = connection.Execute("select 1")
		if err != nil {
			rootCmd.PrintErrf("check connection for target url %v fail with error: %v\n", targetSqlUrl, err)
			return
		}
		rootCmd.Printf("connection established\n")
		// Open reader and do StatementStream
		for _, fileName := range args {
			rootCmd.Printf("process file %v", fileName)
			err := processFile(fileName, sqlDialect, connection, debugLevel)
			if err != nil {
				rootCmd.Println(" - fail")
				rootCmd.Println()
				rootCmd.PrintErrf("Error is %v", err)
				rootCmd.PrintErrln()
			} else {
				rootCmd.Println(" - ok")
			}
		}
	},
}

func init() {
	loadCmd.Flags().StringP("target-sql-connection", "c", "sqlite3://./local.sqlite3", `
Sql url for loading dump file. Examples:

    mysql://user:password@/dbname            // [MySQL, MariaDB, TiDB]
    sqlite3://./local.sqlite3?cache=shared   // [Sqlite3]
    pg://username:password@localhost:5432/database_name    // [PostgresQL]

`)
	loadCmd.Flags().IntP("debug-level", "d", 0, `
Debug level:

	0 no debug messages
	1 show debug messages
	2 show advanced debug messages

`)
	rootCmd.AddCommand(loadCmd)
}

func processFile(fileName string, sqlDialect dialect.SqlDialect, connection sql_connection.SqlConnection, debugLevel int) error {
	respBody, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("file %s open error (%v)", fileName, err)
	}
	defer respBody.Close()

	reader := archive_stream.NewReader(respBody)

	for {
		entry, err := reader.GetNextEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("unable to get next entry (%v)", err)
		}

		if !entry.IsDir() {
			rc, err := entry.Open()
			defer func() {
				if err := rc.Close(); err != nil {
					rootCmd.PrintErrf("close entry reader fail: %s", err)
				}
			}()

			if err != nil {
				return fmt.Errorf("unable to open file: %s", err)
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
					executionError := connection.Execute(statementText)
					if executionError != nil {
						if debugLevel >= 1 {
							rootCmd.PrintErrf("execute sql statement:\n %s \n\nfail: %s\n", statementText, executionError)
						} else {
							rootCmd.PrintErrf("%s\n", executionError)
						}
					}
					statementsCount++
					if debugLevel >= 2 {
						rootCmd.Printf("[%v] processed statements: %v\n", time.Since(lastTime), statementsCount)
					}
					lastTime = time.Now()
				})
		}
	}
	return nil
}
