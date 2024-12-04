package cmd

import (
	"github.com/spf13/cobra"
	"github.com/usalko/prodl/internal/sql_connection"
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

		err = connection.Execute("select 1")
		if err != nil {
			rootCmd.PrintErrf("check connection for target url %v fail with error: %v\n", targetSqlUrl, err)
			return
		}
		rootCmd.Printf("connection established\n")
		for _, fileName := range args {
			rootCmd.Printf("process file %v\n", fileName)
		}

		// 1. Test connection to the database
		// 2. Open reader and do StatementStream
	},
}

func init() {
	loadCmd.Flags().StringP("target-sql-connection", "c", "sqlite3://./local.sqlite3", `
Sql url for loading dump file. Examples:

    mysql://user:password@/dbname            // [MySQL, MariaDB, TiDB]
    sqlite3://./local.sqlite3?cache=shared   // [Sqlite3]
    pg://username:password@localhost:5432/database_name    // [PostgresQL]

`)
	rootCmd.AddCommand(loadCmd)
}
