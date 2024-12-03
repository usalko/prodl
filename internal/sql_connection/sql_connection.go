package sql_connection

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/usalko/prodl/internal/sql_parser/dialect"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type SqlConnection interface {
	Establish(connectionOptions string) error
	Execute(rawSql string) error
}

type MysqlConnection struct {
	db *sql.DB
}

// Establish implements SqlConnection.
func (mysqlConnection MysqlConnection) Establish(connectionOptions string) error {
	db, err := sql.Open("mysql", connectionOptions)
	if err != nil {
		return err
	}
	mysqlConnection.db = db
	return nil
}

// Execute implements SqlConnection.
func (mysqlConnection MysqlConnection) Execute(rawSql string) error {
	_, err := mysqlConnection.db.Exec(rawSql)
	return err
}

type Sqlite3Connection struct {
	db *sql.DB
}

// Establish implements SqlConnection.
func (sqlite3Connection Sqlite3Connection) Establish(connectionOptions string) error {
	db, err := sql.Open("sqlite3", connectionOptions)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1) // Sqlite3 specific
	sqlite3Connection.db = db
	return nil
}

// Execute implements SqlConnection.
func (sqlite3Connection Sqlite3Connection) Execute(rawSql string) error {
	_, err := sqlite3Connection.db.Exec(rawSql)
	return err
}

type PgConnection struct {
	pgxOptions string
}

// Establish implements SqlConnection.
func (pgConnection PgConnection) Establish(connectionOptions string) error {
	pgConnection.pgxOptions = "postgres://" + connectionOptions
	return nil
}

// Execute implements SqlConnection.
func (pgConnection PgConnection) Execute(rawSql string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgConn, err := pgconn.Connect(ctx, pgConnection.pgxOptions)
	if err != nil {
		return err
	}
	defer pgConn.Close(ctx)

	result := pgConn.ExecParams(ctx, rawSql, nil, nil, nil, nil).Read()
	if result.Err != nil {
		return result.Err
	}
	return nil
}

// connection factory
func Connect(sqlDialect dialect.SqlDialect) (SqlConnection, error) {
	switch sqlDialect {
	case dialect.MYSQL:
		return MysqlConnection{}, nil
	case dialect.SQLITE3:
		return Sqlite3Connection{}, nil
	case dialect.PSQL:
		return PgConnection{}, nil
	}
	return nil, fmt.Errorf("can't make connection cause dialect %v not supported for now, please contact with authors", sqlDialect)
}
