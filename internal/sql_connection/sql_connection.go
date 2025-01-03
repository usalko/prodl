package sql_connection

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/usalko/prodl/internal/sql_parser/dialect"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type DbTable struct {
	TableCatalog              string
	TableSchema               string
	TableName                 string
	TableType                 string
	SelfReferencingColumnName string
	ReferenceGeneration       string
	UserDefinedTypeCatalog    string
	UserDefinedTypeSchema     string
	UserDefinedTypeName       string
	IsInsertableInto          string
	IsTyped                   string
	CommitAction              string
}

type DbStructure struct {
	Tables []*DbTable
}

type SqlConnection interface {
	Establish(connectionOptions string) error
	Execute(rawSql string) error
	GetStructure(schemaPattern string, includeSystemTables bool) (*DbStructure, error)
}

type MysqlConnection struct {
	db *sql.DB
}

// Establish implements SqlConnection.
func (mysqlConnection *MysqlConnection) Establish(connectionOptions string) error {
	db, err := sql.Open("mysql", connectionOptions)
	if err != nil {
		return err
	}
	mysqlConnection.db = db
	return nil
}

// Execute implements SqlConnection.
func (mysqlConnection *MysqlConnection) Execute(rawSql string) error {
	_, err := mysqlConnection.db.Exec(rawSql)
	return err
}

// GetStructure implements SqlConnection.
func (mysqlConnection *MysqlConnection) GetStructure(schemaPattern string, includeSystemTables bool) (*DbStructure, error) {
	rows, err := mysqlConnection.db.Query(`SELECT *
FROM INFORMATION_SCHEMA.Tables`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// An album slice to hold data from returned rows.
	result := &DbStructure{
		Tables: make([]*DbTable, 0, 100),
	}

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var table DbTable
		if err := rows.Scan(
			&table.TableCatalog,
			&table.TableSchema,
			&table.TableName,
			&table.TableType,
			&table.SelfReferencingColumnName,
			&table.ReferenceGeneration,
			&table.UserDefinedTypeCatalog,
			&table.UserDefinedTypeSchema,
			&table.UserDefinedTypeName,
			&table.IsInsertableInto,
			&table.IsTyped,
			&table.CommitAction,
		); err != nil {
			return nil, err
		}
		result.Tables = append(result.Tables, &table)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

type Sqlite3Connection struct {
	db *sql.DB
}

// Establish implements SqlConnection.
func (sqlite3Connection *Sqlite3Connection) Establish(connectionOptions string) error {
	db, err := sql.Open("sqlite3", connectionOptions)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1) // Sqlite3 specific
	sqlite3Connection.db = db
	return nil
}

// Execute implements SqlConnection.
func (sqlite3Connection *Sqlite3Connection) Execute(rawSql string) error {
	_, err := sqlite3Connection.db.Exec(rawSql)
	return err
}

// GetStructure implements SqlConnection.
func (sqlite3Connection *Sqlite3Connection) GetStructure(schemaPattern string, includeSystemTables bool) (*DbStructure, error) {
	rows, err := sqlite3Connection.db.Query(`SELECT *
FROM INFORMATION_SCHEMA.Tables`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// An album slice to hold data from returned rows.
	result := &DbStructure{
		Tables: make([]*DbTable, 0, 100),
	}

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var table DbTable
		if err := rows.Scan(
			&table.TableCatalog,
			&table.TableSchema,
			&table.TableName,
			&table.TableType,
			&table.SelfReferencingColumnName,
			&table.ReferenceGeneration,
			&table.UserDefinedTypeCatalog,
			&table.UserDefinedTypeSchema,
			&table.UserDefinedTypeName,
			&table.IsInsertableInto,
			&table.IsTyped,
			&table.CommitAction,
		); err != nil {
			return nil, err
		}
		result.Tables = append(result.Tables, &table)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

type PgConnection struct {
	pgxOptions string
}

// Establish implements SqlConnection.
func (pgConnection *PgConnection) Establish(connectionOptions string) error {
	pgConnection.pgxOptions = "postgres://" + connectionOptions
	return nil
}

// Execute implements SqlConnection.
func (pgConnection *PgConnection) Execute(rawSql string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgConn, err := pgconn.Connect(ctx, pgConnection.pgxOptions)
	if err != nil {
		return err
	}
	defer pgConn.Close(ctx)

	// Recognize COPY FROM STDIN command
	if strings.Contains(rawSql, "COPY") && strings.HasSuffix(rawSql, "\\.") {
		sqlCommandAndData := strings.SplitN(rawSql, "stdin;\n", 2)
		_, err := pgConn.CopyFrom(ctx, strings.NewReader(sqlCommandAndData[1][:len(sqlCommandAndData[1])-2]), sqlCommandAndData[0]+" stdin;")
		if err != nil {
			return err
		}
		return nil
	}
	result := pgConn.ExecParams(ctx, rawSql, nil, nil, nil, nil).Read()
	if result.Err != nil {
		return result.Err
	}
	return nil
}

// Query implements SqlConnection.
func (pgConnection *PgConnection) Query(rawSql string) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgConn, err := pgconn.Connect(ctx, pgConnection.pgxOptions)
	if err != nil {
		return nil, err
	}
	defer pgConn.Close(ctx)

	// Recognize COPY FROM STDIN command
	result := pgConn.ExecParams(ctx, rawSql, nil, nil, nil, nil).Read()
	if result.Err != nil {
		return nil, result.Err
	}
	return &sql.Rows{}, nil
}

// GetStructure implements SqlConnection.
func (pgConnection *PgConnection) GetStructure(schemaPattern string, includeSystemTables bool) (*DbStructure, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgConn, err := pgconn.Connect(ctx, pgConnection.pgxOptions)
	if err != nil {
		return nil, err
	}
	defer pgConn.Close(ctx)

	// An tables slice to hold data from returned rows.
	result := &DbStructure{
		Tables: make([]*DbTable, 0, 100),
	}

	getTablesQuery := `SELECT *
FROM INFORMATION_SCHEMA.Tables`
	if !includeSystemTables {
		getTablesQuery += `
WHERE table_schema <> 'pg_catalog' and table_schema <> 'information_schema'`
	}

	resultReader := pgConn.ExecParams(ctx, getTablesQuery, nil, nil, nil, nil).Read()
	if resultReader.Err != nil {
		return nil, resultReader.Err
	}
	for _, row := range resultReader.Rows {
		table := DbTable{
			TableCatalog:              string(row[0]),
			TableSchema:               string(row[1]),
			TableName:                 string(row[2]),
			TableType:                 string(row[3]),
			SelfReferencingColumnName: string(row[4]),
			ReferenceGeneration:       string(row[5]),
			UserDefinedTypeCatalog:    string(row[6]),
			UserDefinedTypeSchema:     string(row[7]),
			UserDefinedTypeName:       string(row[8]),
			IsInsertableInto:          string(row[9]),
			IsTyped:                   string(row[10]),
			CommitAction:              string(row[11]),
		}
		result.Tables = append(result.Tables, &table)
	}

	return result, nil
}

// connection factory
func Connect(sqlDialect dialect.SqlDialect) (SqlConnection, error) {
	switch sqlDialect {
	case dialect.MYSQL:
		return &MysqlConnection{}, nil
	case dialect.SQLITE3:
		return &Sqlite3Connection{}, nil
	case dialect.PSQL:
		return &PgConnection{}, nil
	}
	return nil, fmt.Errorf("can't make connection cause dialect %v not supported for now, please contact with authors", sqlDialect)
}
