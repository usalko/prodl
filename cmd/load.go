package cmd

import (
	"github.com/spf13/cobra"
	"github.com/usalko/prodl/internal/sql_connection"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "The 'load' subcommand will load dump to the database.",
	Long: `The 'load' subcommand loads a sql dump to the database. For example:

'<cmd> load --to sqlite3://./local.sqlite3 dump-file-name.tar.gz'.`,
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

		// 1. Test connection to the database
		// 2. Open reader and do StatementStream
		rootCmd.Printf("connection established\n")
	},
}

func init() {
	loadCmd.Flags().StringP("target-sql-connection", "c", "sqlite3://./local.sqlite3", `
Sql url for loading dump file. Examples:
[MySQL, MariaDB, TiDB] 	mysql://user:password@/dbname
[Sqlite3]				sqlite3://./local.sqlite3?cache=shared
[PostgresQL]			pg://username:password@localhost:5432/database_name
`)
	rootCmd.AddCommand(loadCmd)
}
