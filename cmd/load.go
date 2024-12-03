package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/mattn/go-sqlite3"
)

func parseTargetSqlUrl(input string) (string, string, error) {
	inputComponents := strings.SplitN(input, "//", 2)
	driverId := strings.Trim(strings.ToLower(inputComponents[0]), ":")
	if len(inputComponents) != 2 {
		return driverId, "", fmt.Errorf("missed connection info in url: %v", input)
	}
	switch driverId {
	case "sqlite3":
		return driverId, "file:" + inputComponents[1], nil
	}
	// db, err := sql.Open("mysql", "user:password@/dbname")
	// db, err := sql.Open("postgres", "username:password@localhost:5432/database_name")
	// db, err := sql.Open("sqlite3", "file:locked.sqlite?cache=shared")
	return driverId, inputComponents[1], nil
}

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "The 'load' subcommand will load dump to the database.",
	Long: `The 'load' subcommand loads a sql dump to the database. For example:

'<cmd> load --to sqlite3://./local.sqlite3 dump-file-name.tar.gz'.`,
	Run: func(cmd *cobra.Command, args []string) {
		targetSqlUrl, _ := cmd.Flags().GetString("target-sql-connection")
		driverId, connectionOptions, err := parseTargetSqlUrl(targetSqlUrl)
		if err != nil {
			rootCmd.PrintErrf("parse target url %v fail with error: %v\n", targetSqlUrl, err)
			return
		}

		db, err := sql.Open(driverId, connectionOptions)
		if err != nil {
			rootCmd.PrintErrf("connection to the database %v fail with error: %v\n", targetSqlUrl, err)
			return
		}
		db.SetMaxOpenConns(1) // Sqlite3 specific
		db.Exec("select 1")
		// 1. Test connection to the database
		// 2. Open reader and do StatementStream
	},
}

func init() {
	loadCmd.Flags().StringP("target-sql-connection", "c", "sqlite3://./local.sqlite3", `
Sql url for loading dump file. Examples:
[MySQL, MariaDB, TiDB] 	mysql://user:password@/dbname
[Sqlite3]				sqlite3://./local.sqlite3?cache=shared
[PostgresQL]			pg://user:password@/dbname
`)
	rootCmd.AddCommand(loadCmd)
}
