package dialect

import (
	"fmt"
	"strings"
)

// All available dialects for parsing

type SqlDialect uint8

const (
	MYSQL   SqlDialect = 1
	PSQL    SqlDialect = 2
	SQLITE3 SqlDialect = 3
)

func (dialect *SqlDialect) String() string {
	switch *dialect {
	case MYSQL:
		return "MYSQL"
	case PSQL:
		return "PSQL"
	case SQLITE3:
		return "SQLITE3"
	}
	return "UNDEFINED"
}

// Parse sql url.
// Example.
//   - input: 	sqlite3://./localfile.sqlite3
//   - output:	SQLITE3, file:./localfile.sqlite3, nil
//
// This output is useful for database/sql package:
//
//	// db, err := sql.Open("mysql", "user:password@/dbname")
//	// db, err := sql.Open("postgres", "username:password@localhost:5432/database_name")
//	// db, err := sql.Open("sqlite3", "file:locked.sqlite?cache=shared")
func (dialect *SqlDialect) ParseUrl(url string) (SqlDialect, string, error) {
	inputComponents := strings.SplitN(url, "//", 2)
	driverId := strings.Trim(strings.ToLower(inputComponents[0]), ":")
	if len(inputComponents) != 2 {
		return 0, "", fmt.Errorf("missed connection info in url: %v", url)
	}
	connectionOptions := inputComponents[1]
	switch driverId {
	case "mysql":
		return MYSQL, connectionOptions, nil
	case "sqlite3":
		return SQLITE3, "file:" + connectionOptions, nil
	case "pg":
		return PSQL, inputComponents[1], nil
	}
	return 0, inputComponents[1], fmt.Errorf("unknown driver identity: %v, for url: %v", driverId, url)
}
