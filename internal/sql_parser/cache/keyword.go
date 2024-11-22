package cache

type Keyword struct {
	Name string
	Id   int
}

func (k *Keyword) match(input []byte) bool {
	if len(input) != len(k.Name) {
		return false
	}
	for i, c := range input {
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		if k.Name[i] != c {
			return false
		}
	}
	return true
}

func (k *Keyword) MatchStr(input string) bool {
	return KeywordASCIIMatch(input, k.Name)
}

func KeywordASCIIMatch(input string, expected string) bool {
	if len(input) != len(expected) {
		return false
	}
	for i := 0; i < len(input); i++ {
		c := input[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		if expected[i] != c {
			return false
		}
	}
	return true
}
