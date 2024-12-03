package dialect

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
