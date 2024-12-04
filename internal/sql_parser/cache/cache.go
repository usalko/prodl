package cache

import "github.com/usalko/prodl/internal/sql_parser/dialect"

// KeywordLookupTable is a perfect hash map that maps **case insensitive** keyword names to their ids
var KeywordLookupTables map[dialect.SqlDialect]*CaseInsensitiveTable = map[dialect.SqlDialect]*CaseInsensitiveTable{}

func KeywordLookup(s string, sqlDialect dialect.SqlDialect) (int, bool) {
	lookupTable, ok := KeywordLookupTables[sqlDialect]
	if !ok || lookupTable == nil {
		return 0, false
	}
	return lookupTable.LookupString(s)
}

type CaseInsensitiveTable struct {
	Hashes map[uint64]Keyword
}

func (cit *CaseInsensitiveTable) LookupString(name string) (int, bool) {
	hash := Fnv1aIstr(Offset64, name)
	if candidate, ok := cit.Hashes[hash]; ok {
		return candidate.Id, candidate.MatchStr(name)
	}
	return 0, false
}

func (cit *CaseInsensitiveTable) Lookup(name []byte) (int, bool) {
	hash := Fnv1aI(Offset64, name)
	if candidate, ok := cit.Hashes[hash]; ok {
		return candidate.Id, candidate.match(name)
	}
	return 0, false
}

const Offset64 = uint64(14695981039346656037)
const Prime64 = uint64(1099511628211)

func Fnv1aI(h uint64, s []byte) uint64 {
	for _, c := range s {
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		h = (h ^ uint64(c)) * Prime64
	}
	return h
}

func Fnv1aIstr(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		h = (h ^ uint64(c)) * Prime64
	}
	return h
}
