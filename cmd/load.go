package cmd

import (
	"database/sql"

	"github.com/spf13/cobra"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "The 'load' subcommand will load dump to the database.",
	Long: `The 'load' subcommand loads a sql dump to the database. For example:

'<cmd> load --to sqlite://local.sqlite3 dump-file-name.tar.gz'.`,
	Run: func(cmd *cobra.Command, args []string) {
		targetSqlUrl, _ := cmd.Flags().GetString("target-sql-connection")

		db, err := sql.Open("sqlite3", targetSqlUrl)
		if err != nil {
			rootCmd.PrintErrf("Connection to the dtabase %v fail with error %v", targetSqlUrl, err)
			return
		}
		db.Exec("select 1")
		// 1. Test connection to the database
		// 2. Open reader and do StatementStream
		println(targetSqlUrl)
	},
}

func init() {
	loadCmd.Flags().StringP("target-sql-connection", "c", "sqlite://./local.sqlite3", "Sql url for loading dump file")
	rootCmd.AddCommand(loadCmd)
}
