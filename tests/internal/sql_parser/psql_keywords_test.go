package sql_parser

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/usalko/prodl/internal/sql_parser"
	"github.com/usalko/prodl/internal/sql_parser/cache"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
	"github.com/usalko/prodl/internal/sql_parser/psql"
)

func TestKeywordTable(t *testing.T) {
	for _, kw := range psql.GetKeywords() {
		lookup, ok := cache.KeywordLookup(kw.Name, dialect.PSQL)
		require.Truef(t, ok, "keyword %q failed to match", kw.Name)
		require.Equalf(t, lookup, kw.Id, "keyword %q matched to %d (expected %d)", kw.Name, lookup, kw.Id)
	}
}

func TestCompatibility(t *testing.T) {
	file, err := os.Open(path.Join("test_data", "psql_keywords.txt"))
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	skipStep := 2
	for scanner.Scan() {
		if skipStep != 0 {
			skipStep--
			continue
		}

		afterSplit := strings.SplitN(scanner.Text(), "\t", 3)
		word, reserved := strings.ReplaceAll(afterSplit[0], "\u200b", ""), strings.Trim(afterSplit[1], " ") != ""
		if reserved {
			word = "\"" + word + "\""
		}
		sql := fmt.Sprintf("create table %s(c1 int)", word)
		_, err := sql_parser.ParseStrictDDL(sql, dialect.PSQL)
		if err != nil {
			t.Errorf("%s is not compatible with psql", word)
		}
	}
}
