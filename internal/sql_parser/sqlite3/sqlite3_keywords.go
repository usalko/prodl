/*
Copyright 2021 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sqlite3

import (
	"fmt"
	"strings"

	"github.com/usalko/prodl/internal/sql_parser/cache"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
)

// keywords is a table of sqlite3 keywords that fall into two categories:
// 1) keywords considered reserved by PostgresQL
// 2) keywords for us to handle specially in sqlite3.y
//
// Those marked as UNUSED are likely reserved keywords. We add them here so that
// when rewriting queries we can properly backtick quote them so they don't cause issues
//
// NOTE: If you add new keywords, add them also to the reserved_keywords or
// non_reserved_keywords grammar in sql.y -- this will allow the keyword to be used
// in identifiers. See the docs for each grammar to determine which one to put it into.
var keywords = []cache.Keyword{
	{Name: "abort", Id: ABORT},
	{Name: "action", Id: ACTION},
	{Name: "add", Id: ADD},
	{Name: "after", Id: AFTER},
	{Name: "all", Id: ALL},
	{Name: "alter", Id: ALTER},
	{Name: "always", Id: ALWAYS},
	{Name: "analyze", Id: ANALYZE},
	{Name: "and", Id: AND},
	{Name: "as", Id: AS},
	{Name: "asc", Id: ASC},
	{Name: "attach", Id: ATTACH},
	{Name: "autoincrement", Id: AUTOINCREMENT},
	{Name: "before", Id: BEFORE},
	{Name: "begin", Id: BEGIN},
	{Name: "between", Id: BETWEEN},
	{Name: "by", Id: BY},
	{Name: "cascade", Id: CASCADE},
	{Name: "case", Id: CASE},
	{Name: "cast", Id: CAST},
	{Name: "check", Id: CHECK},
	{Name: "collate", Id: COLLATE},
	{Name: "column", Id: COLUMN},
	{Name: "commit", Id: COMMIT},
	{Name: "conflict", Id: CONFLICT},
	{Name: "constraint", Id: CONSTRAINT},
	{Name: "create", Id: CREATE},
	{Name: "cross", Id: CROSS},
	{Name: "current", Id: CURRENT},
	{Name: "current_date", Id: CURRENT_DATE},
	{Name: "current_time", Id: CURRENT_TIME},
	{Name: "current_timestamp", Id: CURRENT_TIMESTAMP},
	{Name: "database", Id: DATABASE},
	{Name: "default", Id: DEFAULT},
	{Name: "deferrable", Id: DEFERRABLE},
	{Name: "deferred", Id: DEFERRED},
	{Name: "delete", Id: DELETE},
	{Name: "desc", Id: DESC},
	{Name: "detach", Id: DETACH},
	{Name: "distinct", Id: DISTINCT},
	{Name: "do", Id: DO},
	{Name: "drop", Id: DROP},
	{Name: "each", Id: EACH},
	{Name: "else", Id: ELSE},
	{Name: "end", Id: END},
	{Name: "escape", Id: ESCAPE},
	{Name: "except", Id: EXCEPT},
	{Name: "exclude", Id: EXCLUDE},
	{Name: "exclusive", Id: EXCLUSIVE},
	{Name: "exists", Id: EXISTS},
	{Name: "explain", Id: EXPLAIN},
	{Name: "fail", Id: FAIL},
	{Name: "filter", Id: FILTER},
	{Name: "first", Id: FIRST},
	{Name: "following", Id: FOLLOWING},
	{Name: "for", Id: FOR},
	{Name: "foreign", Id: FOREIGN},
	{Name: "from", Id: FROM},
	{Name: "full", Id: FULL},
	{Name: "generated", Id: GENERATED},
	{Name: "glob", Id: GLOB},
	{Name: "group", Id: GROUP},
	{Name: "groups", Id: GROUPS},
	{Name: "having", Id: HAVING},
	{Name: "if", Id: IF},
	{Name: "ignore", Id: IGNORE},
	{Name: "identity", Id: IDENTITY},
	{Name: "immediate", Id: IMMEDIATE},
	{Name: "in", Id: IN},
	{Name: "index", Id: INDEX},
	{Name: "indexed", Id: INDEXED},
	{Name: "initially", Id: INITIALLY},
	{Name: "inner", Id: INNER},
	{Name: "insert", Id: INSERT},
	{Name: "instead", Id: INSTEAD},
	{Name: "intersect", Id: INTERSECT},
	{Name: "into", Id: INTO},
	{Name: "is", Id: IS},
	{Name: "isnull", Id: ISNULL},
	{Name: "join", Id: JOIN},
	{Name: "key", Id: KEY},
	{Name: "last", Id: LAST},
	{Name: "left", Id: LEFT},
	{Name: "like", Id: LIKE},
	{Name: "limit", Id: LIMIT},
	{Name: "match", Id: MATCH},
	{Name: "materialized", Id: MATERIALIZED},
	{Name: "natural", Id: NATURAL},
	{Name: "no", Id: NO},
	{Name: "not", Id: NOT},
	{Name: "nothing", Id: NOTHING},
	{Name: "notnull", Id: NOTNULL},
	{Name: "null", Id: NULL},
	{Name: "nulls", Id: NULLS},
	{Name: "of", Id: OF},
	{Name: "offset", Id: OFFSET},
	{Name: "on", Id: ON},
	{Name: "or", Id: OR},
	{Name: "order", Id: ORDER},
	{Name: "others", Id: OTHERS},
	{Name: "outer", Id: OUTER},
	{Name: "over", Id: OVER},
	{Name: "partition", Id: PARTITION},
	{Name: "plan", Id: PLAN},
	{Name: "pragma", Id: PRAGMA},
	{Name: "preceding", Id: PRECEDING},
	{Name: "primary", Id: PRIMARY},
	{Name: "query", Id: QUERY},
	{Name: "raise", Id: RAISE},
	{Name: "range", Id: RANGE},
	{Name: "recursive", Id: RECURSIVE},
	{Name: "references", Id: REFERENCES},
	{Name: "regexp", Id: REGEXP},
	{Name: "reindex", Id: REINDEX},
	{Name: "release", Id: RELEASE},
	{Name: "rename", Id: RENAME},
	{Name: "replace", Id: REPLACE},
	{Name: "restrict", Id: RESTRICT},
	{Name: "returning", Id: RETURNING},
	{Name: "right", Id: RIGHT},
	{Name: "rollback", Id: ROLLBACK},
	{Name: "row", Id: ROW},
	{Name: "rows", Id: ROWS},
	{Name: "savepoint", Id: SAVEPOINT},
	{Name: "select", Id: SELECT},
	{Name: "set", Id: SET},
	{Name: "table", Id: TABLE},
	{Name: "temp", Id: TEMP},
	{Name: "temporary", Id: TEMPORARY},
	{Name: "then", Id: THEN},
	{Name: "ties", Id: TIES},
	{Name: "to", Id: TO},
	{Name: "transaction", Id: TRANSACTION},
	{Name: "trigger", Id: TRIGGER},
	{Name: "unbounded", Id: UNBOUNDED},
	{Name: "union", Id: UNION},
	{Name: "unique", Id: UNIQUE},
	{Name: "update", Id: UPDATE},
	{Name: "using", Id: USING},
	{Name: "vacuum", Id: VACUUM},
	{Name: "values", Id: VALUES},
	{Name: "view", Id: VIEW},
	{Name: "virtual", Id: VIRTUAL},
	{Name: "when", Id: WHEN},
	{Name: "where", Id: WHERE},
	{Name: "window", Id: WINDOW},
	{Name: "with", Id: WITH},
	{Name: "without", Id: WITHOUT},
}

func GetKeywords() []cache.Keyword {
	result := make([]cache.Keyword, len(keywords))
	for i, keyword := range keywords {
		result[i] = cache.Keyword(keyword)
	}
	return result
}

// keywordStrings contains the reverse mapping of token to keyword strings
var keywordStrings = map[int]string{}

func buildCaseInsensitiveTable(keywords []cache.Keyword) *cache.CaseInsensitiveTable {
	table := &cache.CaseInsensitiveTable{
		Hashes: make(map[uint64]cache.Keyword, len(keywords)),
	}

	for _, kw := range keywords {
		hash := cache.Fnv1aIstr(cache.Offset64, kw.Name)
		if _, exists := table.Hashes[hash]; exists {
			panic("collision in caseInsensitiveTable")
		}
		table.Hashes[hash] = kw
	}
	return table
}

func init() {
	for _, kw := range keywords {
		if kw.Id == UNUSED {
			continue
		}
		if kw.Name != strings.ToLower(kw.Name) {
			panic(fmt.Sprintf("keyword %q must be lowercase in table", kw.Name))
		}
		keywordStrings[kw.Id] = kw.Name
	}

	cache.KeywordLookupTables[dialect.SQLITE3] = buildCaseInsensitiveTable(keywords)
}

// KeywordString returns the string corresponding to the given keyword
func KeywordString(id int) string {
	str, ok := keywordStrings[id]
	if !ok {
		return ""
	}
	return str
}
