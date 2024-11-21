package dialect

// All available dialects for parsing

type SqlDialect uint8

const (
	MYSQL SqlDialect = 1
	PSQL  SqlDialect = 2
)
