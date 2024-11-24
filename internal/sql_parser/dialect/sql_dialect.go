package dialect

// All available dialects for parsing

type SqlDialect uint8

const (
	MYSQL SqlDialect = 1
	PSQL  SqlDialect = 2
)

func (dialect *SqlDialect) String() string {
	switch *dialect {
	case MYSQL:
		return "MYSQL"
	case PSQL:
		return "PSQL"
	}
	return "UNDEFINED"
}
