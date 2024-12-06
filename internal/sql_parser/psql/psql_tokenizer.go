/*
Copyright 2019 The Vitess Authors.

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

package psql

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/usalko/prodl/internal/sql_parser/ast"
	"github.com/usalko/prodl/internal/sql_parser/cache"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
	"github.com/usalko/prodl/internal/sql_parser/tokenizer"
	"github.com/usalko/prodl/internal/sql_types"
)

// PsqlTokenizer is the struct used to generate SQL
// tokens for the parser.
type PsqlTokenizer struct {
	AllowComments       bool
	SkipSpecialComments bool
	SkipToEnd           bool
	LastError           error
	ParseTree           ast.Statement
	BindVars            map[string]struct{}

	lastToken      string
	posVarIndex    int
	partialDDL     ast.Statement
	nesting        int
	multi          bool
	specialComment *PsqlTokenizer
	leftContext    tokenizer.CyclicBuffer

	Pos int
	buf string
}

// SetSkipSpecialComments implements tokenizer.Tokenizer.
func (tzr *PsqlTokenizer) SetSkipSpecialComments(skip bool) {
	tzr.SkipSpecialComments = skip
}

// GetBindVars implements tokenizer.Tokenizer.
func (tzr *PsqlTokenizer) GetBindVars() ast.BindVars {
	return tzr.BindVars
}

// GetLastError implements tokenizer.Tokenizer.
func (tzr *PsqlTokenizer) GetLastError() error {
	return tzr.LastError
}

// GetPos implements tokenizer.Tokenizer.
func (tzr *PsqlTokenizer) GetPos() int {
	return tzr.Pos
}

// SetMulti implements tokenizer.Tokenizer.
func (tzr *PsqlTokenizer) SetMulti(multi bool) {
	tzr.multi = multi
}

// GetIdToken implements sql_parser.Tokenizer.
func (tzr *PsqlTokenizer) GetIdToken() int {
	return ID
}

// GetKeywordString implements sql_parser.Tokenizer.
func (tzr *PsqlTokenizer) GetKeywordString(token int) string {
	return KeywordString(token)
}

// GetParseTree implements sql_parser.Tokenizer.
func (tzr *PsqlTokenizer) GetParseTree() ast.Statement {
	return tzr.ParseTree
}

// GetPartialDDL implements sql_parser.Tokenizer.
func (tzr *PsqlTokenizer) GetPartialDDL() ast.Statement {
	return tzr.partialDDL
}

// BindVar implements sql_parser.Tokenizer.
func (tzr *PsqlTokenizer) BindVar(bvar string, value struct{}) {
	tzr.BindVars[bvar] = value
}

// DecNesting implements sql_parser.Tokenizer.
func (tzr *PsqlTokenizer) DecNesting() {
	tzr.nesting--
}

// GetNesting implements sql_parser.Tokenizer.
func (tzr *PsqlTokenizer) GetNesting() int {
	return tzr.nesting
}

// IncNesting implements sql_parser.Tokenizer.
func (tzr *PsqlTokenizer) IncNesting() {
	tzr.nesting++
}

// SetAllowComments implements sql_parser.Tokenizer.
func (tzr *PsqlTokenizer) SetAllowComments(allow bool) {
	tzr.AllowComments = allow
}

// SetParseTree implements sql_parser.Tokenizer.
func (tzr *PsqlTokenizer) SetParseTree(stmt ast.Statement) {
	tzr.ParseTree = stmt
}

// SetPartialDDL implements sql_parser.Tokenizer.
func (tzr *PsqlTokenizer) SetPartialDDL(node ast.Statement) {
	tzr.partialDDL = node
}

// SetSkipToEnd implements sql_parser.Tokenizer.
func (tzr *PsqlTokenizer) SetSkipToEnd(skip bool) {
	tzr.SkipToEnd = skip
}

// parserPool is a pool for parser objects.
var parserPool = sync.Pool{
	New: func() any {
		return &psqParserImpl{}
	},
}

// zeroParser is a zero-initialized parser to help reinitialize the parser for pooling.
var zeroParser psqParserImpl

// psqParsePooled is a wrapper around psqParse that pools the parser objects. There isn't a
// particularly good reason to use psqParse directly, since it immediately discards its parser.
//
// N.B: Parser pooling means that you CANNOT take references directly to parse stack variables (e.g.
// $$ = &$4) in sql.y rules. You must instead add an intermediate reference like so:
//
//	showCollationFilterOpt := $4
//	$$ = &Show{Type: string($2), ShowCollationFilterOpt: &showCollationFilterOpt}
func ParsePooled(lexer tokenizer.Tokenizer) int {
	parser := parserPool.Get().(*psqParserImpl)
	defer func() {
		*parser = zeroParser
		parserPool.Put(parser)
	}()
	return parser.Parse(lexer.(psqLexer))
}

func Parse(lexer tokenizer.Tokenizer) int {
	return psqParse(lexer.(psqLexer))
}

// PSQLServerVersion is what Vitess will present as it's version during the connection handshake,
// and as the value to the @@version system variable. If nothing is provided, Vitess will report itself as
// a specific PSQL version with the vitess version appended to it
var PSQLServerVersion = flag.String("psql_server_version", "", "PSQL server version to advertise.")

// NewPsqlStringTokenizer creates a new Tokenizer for the
// sql string.
func NewPsqlStringTokenizer(sql string) *PsqlTokenizer {

	return &PsqlTokenizer{
		buf:         sql,
		BindVars:    make(map[string]struct{}),
		leftContext: *tokenizer.NewCyclicBuffer(10),
	}
}

// Lex returns the next token form the Tokenizer.
// This function is used by go yacc.
func (tzr *PsqlTokenizer) Lex(lval *psqSymType) int {
	if tzr.SkipToEnd {
		return tzr.skipStatement()
	}

	typ, val := tzr.Scan()
	for typ == COMMENT {
		if tzr.AllowComments || val == "COMMENT" {
			break
		}
		typ, val = tzr.Scan()
	}
	// COPY command omit rest of DATA from stdin
	if typ == ';' && lval.str == "stdin" {
		tzr.SkipToEnd = true
		return 0
	}
	if typ == 0 || typ == ';' || typ == LEX_ERROR {
		// If encounter end of statement or invalid token,
		// we should not accept partially parsed DDLs. They
		// should instead result in parser errors. See the
		// Parse function to see how this is handled.
		tzr.partialDDL = nil
	}
	lval.str = val
	tzr.lastToken = val
	return typ
}

// PositionedErr holds context related to parser errors
type PositionedErr struct {
	Err  string
	Pos  int
	Near string
}

func (p PositionedErr) Error() string {
	if p.Near != "" {
		return fmt.Sprintf("%s at position %v near '%s'", p.Err, p.Pos, p.Near)
	}
	return fmt.Sprintf("%s at position %v", p.Err, p.Pos)
}

// Error is called by go yacc if there's a parsing error.
func (tzr *PsqlTokenizer) Error(err string) {
	tzr.LastError = PositionedErr{Err: err, Pos: tzr.Pos + 1, Near: tzr.lastToken}

	// Try and re-sync to the next statement
	tzr.skipStatement()
}

// Scan scans the tokenizer for the next token and returns
// the token type and an optional value.
func (tzr *PsqlTokenizer) Scan() (int, string) {
	if tzr.specialComment != nil {
		// Enter specialComment scan mode.
		// for scanning such kind of comment: /*! PSQL-specific code */
		specialComment := tzr.specialComment
		tok, val := specialComment.Scan()
		if tok != 0 {
			// return the specialComment scan result as the result
			return tok, val
		}
		// leave specialComment scan mode after all stream consumed.
		tzr.specialComment = nil
	}

	tzr.SkipBlank()
	switch ch := tzr.leftContext.Put(rune(tzr.Cur())); {
	case ch == '@':
		tokenID := AT_ID
		tzr.Skip(1)
		if tzr.Cur() == '@' {
			tokenID = AT_AT_ID
			tzr.Skip(1)
		}
		var tID int
		var tBytes string
		if tzr.Cur() == '`' {
			tzr.Skip(1)
			tID, tBytes = tzr.scanLiteralIdentifier()
		} else if tzr.Cur() == tokenizer.EofChar {
			return LEX_ERROR, ""
		} else {
			tID, tBytes = tzr.scanIdentifier(true)
		}
		if tID == LEX_ERROR {
			return tID, ""
		}
		return tokenID, tBytes
	case isLetter(ch):
		if ch == 'X' || ch == 'x' {
			if tzr.Peek(1) == '\'' {
				tzr.Skip(2)
				return tzr.scanHex()
			}
		}
		if ch == 'B' || ch == 'b' {
			if tzr.Peek(1) == '\'' {
				tzr.Skip(2)
				return tzr.scanBitLiteral()
			}
		}
		// N\'literal' is used to create a string in the national character set
		if ch == 'N' || ch == 'n' {
			nxt := tzr.Peek(1)
			if nxt == '\'' || nxt == '"' {
				tzr.Skip(2)
				return tzr.scanString(nxt, NCHAR_STRING)
			}
		}
		return tzr.scanIdentifier(false)
	case isDigit(ch):
		return tzr.scanNumber()
	case ch == ':':
		return tzr.scanBindVar()
	case ch == ';':
		if tzr.multi {
			// In multi mode, ';' is treated as EOF. So, we don't advance.
			// Repeated calls to Scan will keep returning 0 until ParseNext
			// forces the advance.
			return 0, ""
		}
		if tzr.leftContext.Has("stdin") {
			tzr.scanEndDataMark()
			return ';', ""
		}
		tzr.Skip(1)
		return ';', ""
	case ch == tokenizer.EofChar:
		return 0, ""
	default:
		if ch == '.' && isDigit(tzr.Peek(1)) {
			return tzr.scanNumber()
		}

		tzr.Skip(1)
		switch ch {
		case '=', ',', '(', ')', '+', '*', '%', '^', '~':
			return int(ch), ""
		case '&':
			if tzr.Cur() == '&' {
				tzr.Skip(1)
				return AND, ""
			}
			return int(ch), ""
		case '|':
			if tzr.Cur() == '|' {
				tzr.Skip(1)
				return OR, ""
			}
			return int(ch), ""
		case '?':
			tzr.posVarIndex++
			buf := make([]byte, 0, 8)
			buf = append(buf, ":v"...)
			buf = strconv.AppendInt(buf, int64(tzr.posVarIndex), 10)
			return VALUE_ARG, string(buf)
		case '.':
			return int(ch), ""
		case '/':
			switch tzr.Cur() {
			case '/':
				tzr.Skip(1)
				return tzr.scanCommentType1(2)
			case '*':
				tzr.Skip(1)
				if tzr.Cur() == '!' && !tzr.SkipSpecialComments {
					tzr.Skip(1)
					return tzr.scanPSQLSpecificComment()
				}
				return tzr.scanCommentType2()
			default:
				return int(ch), ""
			}
		case '#':
			return tzr.scanCommentType1(1)
		case '-':
			switch tzr.Cur() {
			case '-':
				nextChar := tzr.Peek(1)
				if nextChar == ' ' || nextChar == '\n' || nextChar == '\t' || nextChar == '\r' || nextChar == tokenizer.EofChar {
					tzr.Skip(1)
					return tzr.scanCommentType1(2)
				}
			case '>':
				tzr.Skip(1)
				if tzr.Cur() == '>' {
					tzr.Skip(1)
					return JSON_UNQUOTE_EXTRACT_OP, ""
				}
				return JSON_EXTRACT_OP, ""
			}
			return int(ch), ""
		case '<':
			switch tzr.Cur() {
			case '>':
				tzr.Skip(1)
				return NE, ""
			case '<':
				tzr.Skip(1)
				return SHIFT_LEFT, ""
			case '=':
				tzr.Skip(1)
				switch tzr.Cur() {
				case '>':
					tzr.Skip(1)
					return NULL_SAFE_EQUAL, ""
				default:
					return LE, ""
				}
			default:
				return int(ch), ""
			}
		case '>':
			switch tzr.Cur() {
			case '=':
				tzr.Skip(1)
				return GE, ""
			case '>':
				tzr.Skip(1)
				return SHIFT_RIGHT, ""
			default:
				return int(ch), ""
			}
		case '!':
			if tzr.Cur() == '=' {
				tzr.Skip(1)
				return NE, ""
			}
			return int(ch), ""
		case '\'', '"':
			return tzr.scanString(ch, STRING)
		case '`':
			return tzr.scanLiteralIdentifier()
		default:
			return LEX_ERROR, string(byte(ch))
		}
	}
}

// skipStatement scans until end of statement.
func (tzr *PsqlTokenizer) skipStatement() int {
	tzr.SkipToEnd = false
	for {
		typ, _ := tzr.Scan()
		if typ == 0 || typ == ';' || typ == LEX_ERROR {
			return typ
		}
	}
}

// SkipBlank skips the cursor while it finds whitespace
func (tzr *PsqlTokenizer) SkipBlank() {
	ch := tzr.Cur()
	for ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' {
		tzr.Skip(1)
		ch = tzr.Cur()
	}
}

// scanIdentifier scans a language keyword or @-encased variable
func (tzr *PsqlTokenizer) scanIdentifier(isVariable bool) (int, string) {
	start := tzr.Pos
	tzr.Skip(1)

	for {
		ch := tzr.leftContext.Put(rune(tzr.Cur()))
		if !isLetter(ch) && !isDigit(ch) && !(isVariable && isCarat(ch)) {
			break
		}
		tzr.Skip(1)
	}
	keywordName := tzr.buf[start:tzr.Pos]
	if keywordID, found := cache.KeywordLookup(keywordName, dialect.PSQL); found {
		return keywordID, keywordName
	}
	// dual must always be case-insensitive
	if cache.KeywordASCIIMatch(keywordName, "dual") {
		return ID, "dual"
	}
	return ID, keywordName
}

// scanHex scans a hex numeral; assumes x' or X' has already been scanned
func (tzr *PsqlTokenizer) scanHex() (int, string) {
	start := tzr.Pos
	tzr.scanMantissa(16)
	hex := tzr.buf[start:tzr.Pos]
	if tzr.Cur() != '\'' {
		return LEX_ERROR, hex
	}
	tzr.Skip(1)
	if len(hex)%2 != 0 {
		return LEX_ERROR, hex
	}
	return HEX, hex
}

// scanBitLiteral scans a binary numeric literal; assumes b' or B' has already been scanned
func (tzr *PsqlTokenizer) scanBitLiteral() (int, string) {
	start := tzr.Pos
	tzr.scanMantissa(2)
	bit := tzr.buf[start:tzr.Pos]
	if tzr.Cur() != '\'' {
		return LEX_ERROR, bit
	}
	tzr.Skip(1)
	return BIT_LITERAL, bit
}

// scanLiteralIdentifierSlow scans an identifier surrounded by backticks which may
// contain escape sequences instead of it. This method is only called from
// scanLiteralIdentifier once the first escape sequence is found in the identifier.
// The provided `buf` contains the contents of the identifier that have been scanned
// so far.
func (tzr *PsqlTokenizer) scanLiteralIdentifierSlow(buf *strings.Builder) (int, string) {
	backTickSeen := true
	for {
		if backTickSeen {
			if tzr.Cur() != '`' {
				break
			}
			backTickSeen = false
			buf.WriteByte('`')
			tzr.Skip(1)
			continue
		}
		// The previous char was not a backtick.
		switch tzr.Cur() {
		case '`':
			backTickSeen = true
		case tokenizer.EofChar:
			// Premature EOF.
			return LEX_ERROR, buf.String()
		default:
			buf.WriteByte(byte(tzr.Cur()))
			// keep scanning
		}
		tzr.Skip(1)
	}
	return ID, buf.String()
}

// scanLiteralIdentifier scans an identifier enclosed by backticks. If the identifier
// is a simple literal, it'll be returned as a slice of the input buffer. If the identifier
// contains escape sequences, this function will fall back to scanLiteralIdentifierSlow
func (tzr *PsqlTokenizer) scanLiteralIdentifier() (int, string) {
	start := tzr.Pos
	for {
		switch tzr.Cur() {
		case '`':
			if tzr.Peek(1) != '`' {
				if tzr.Pos == start {
					return LEX_ERROR, ""
				}
				tzr.Skip(1)
				return ID, tzr.buf[start : tzr.Pos-1]
			}

			var buf strings.Builder
			buf.WriteString(tzr.buf[start:tzr.Pos])
			tzr.Skip(1)
			return tzr.scanLiteralIdentifierSlow(&buf)
		case tokenizer.EofChar:
			// Premature EOF.
			return LEX_ERROR, tzr.buf[start:tzr.Pos]
		default:
			tzr.Skip(1)
		}
	}
}

// scanBindVar scans a bind variable; assumes a ':' has been scanned right before
func (tzr *PsqlTokenizer) scanBindVar() (int, string) {
	start := tzr.Pos
	token := VALUE_ARG

	tzr.Skip(1)
	if tzr.Cur() == ':' {
		token = LIST_ARG
		tzr.Skip(1)
	}
	if !isLetter(tzr.Cur()) {
		return LEX_ERROR, tzr.buf[start:tzr.Pos]
	}
	for {
		ch := tzr.Cur()
		if !isLetter(ch) && !isDigit(ch) && ch != '.' {
			break
		}
		tzr.Skip(1)
	}
	return token, tzr.buf[start:tzr.Pos]
}

// scanEndDataMark scans a mark for end input data "\\."
func (tzr *PsqlTokenizer) scanEndDataMark() (int, string) {

	tzr.Skip(1)
	start := tzr.Pos
	for {
		ch := tzr.Cur()
		if ch == '\\' {
			tzr.Skip(1)
			if tzr.Cur() == '.' {
				tzr.Skip(1)
				break
			}
		}
		if ch == tokenizer.EofChar {
			break
		}
		tzr.Skip(1)
	}
	return '.', tzr.buf[start : tzr.Pos-1]
}

// scanMantissa scans a sequence of numeric characters with the same base.
// This is a helper function only called from the numeric scanners
func (tzr *PsqlTokenizer) scanMantissa(base int) {
	for digitVal(tzr.Cur()) < base {
		tzr.Skip(1)
	}
}

// scanNumber scans any SQL numeric literal, either floating point or integer
func (tzr *PsqlTokenizer) scanNumber() (int, string) {
	start := tzr.Pos
	token := INTEGRAL

	if tzr.Cur() == '.' {
		token = DECIMAL
		tzr.Skip(1)
		tzr.scanMantissa(10)
		goto exponent
	}

	// 0x construct.
	if tzr.Cur() == '0' {
		tzr.Skip(1)
		if tzr.Cur() == 'x' || tzr.Cur() == 'X' {
			token = HEXNUM
			tzr.Skip(1)
			tzr.scanMantissa(16)
			goto exit
		}
	}

	tzr.scanMantissa(10)

	if tzr.Cur() == '.' {
		token = DECIMAL
		tzr.Skip(1)
		tzr.scanMantissa(10)
	}

exponent:
	if tzr.Cur() == 'e' || tzr.Cur() == 'E' {
		token = FLOAT
		tzr.Skip(1)
		if tzr.Cur() == '+' || tzr.Cur() == '-' {
			tzr.Skip(1)
		}
		tzr.scanMantissa(10)
	}

exit:
	if isLetter(tzr.Cur()) {
		// A letter cannot immediately follow a float number.
		if token == FLOAT || token == DECIMAL {
			return LEX_ERROR, tzr.buf[start:tzr.Pos]
		}
		// A letter seen after a few numbers means that we should parse this
		// as an identifier and not a number.
		for {
			ch := tzr.Cur()
			if !isLetter(ch) && !isDigit(ch) {
				break
			}
			tzr.Skip(1)
		}
		return ID, tzr.buf[start:tzr.Pos]
	}

	return token, tzr.buf[start:tzr.Pos]
}

// scanString scans a string surrounded by the given `delim`, which can be
// either single or double quotes. Assumes that the given delimiter has just
// been scanned. If the skin contains any escape sequences, this function
// will fall back to scanStringSlow
func (tzr *PsqlTokenizer) scanString(delim rune, typ int) (int, string) {
	start := tzr.Pos

	for {
		switch tzr.Cur() {
		case delim:
			if tzr.Peek(1) != delim {
				tzr.Skip(1)
				return typ, tzr.buf[start : tzr.Pos-1]
			}
			fallthrough

		case '\\':
			var buffer strings.Builder
			buffer.WriteString(tzr.buf[start:tzr.Pos])
			return tzr.scanStringSlow(&buffer, delim, typ)

		case tokenizer.EofChar:
			return LEX_ERROR, tzr.buf[start:tzr.Pos]
		}

		tzr.Skip(1)
	}
}

// scanString scans a string surrounded by the given `delim` and containing escape
// sequencse. The given `buffer` contains the contents of the string that have
// been scanned so far.
func (tzr *PsqlTokenizer) scanStringSlow(buffer *strings.Builder, delim rune, typ int) (int, string) {
	for {
		ch := tzr.Cur()
		if ch == tokenizer.EofChar {
			// Unterminated string.
			return LEX_ERROR, buffer.String()
		}

		if ch != delim && ch != '\\' {
			// Scan ahead to the next interesting character.
			start := tzr.Pos
			for ; tzr.Pos < len(tzr.buf); tzr.Pos++ {
				ch = rune(tzr.buf[tzr.Pos])
				if ch == delim || ch == '\\' {
					break
				}
			}

			buffer.WriteString(tzr.buf[start:tzr.Pos])
			if tzr.Pos >= len(tzr.buf) {
				// Reached the end of the buffer without finding a delim or
				// escape character.
				tzr.Skip(1)
				continue
			}
		}
		tzr.Skip(1) // Read one past the delim or escape character.

		if ch == '\\' {
			if tzr.Cur() == tokenizer.EofChar {
				// String terminates mid escape character.
				return LEX_ERROR, buffer.String()
			}
			if decodedChar := sql_types.SQLDecodeMap[byte(tzr.Cur())]; decodedChar == sql_types.DontEscape {
				ch = tzr.Cur()
			} else {
				ch = rune(decodedChar)
			}
		} else if ch == delim && tzr.Cur() != delim {
			// Correctly terminated string, which is not a double delim.
			break
		}

		buffer.WriteByte(byte(ch))
		tzr.Skip(1)
	}

	return typ, buffer.String()
}

// scanCommentType1 scans a SQL line-comment, which is applied until the end
// of the line. The given prefix length varies based on whether the comment
// is started with '//', '--' or '#'.
func (tzr *PsqlTokenizer) scanCommentType1(prefixLen int) (int, string) {
	start := tzr.Pos - prefixLen
	for tzr.Cur() != tokenizer.EofChar {
		if tzr.Cur() == '\n' {
			tzr.Skip(1)
			break
		}
		tzr.Skip(1)
	}
	return COMMENT, tzr.buf[start:tzr.Pos]
}

// scanCommentType2 scans a '/*' delimited comment; assumes the opening
// prefix has already been scanned
func (tzr *PsqlTokenizer) scanCommentType2() (int, string) {
	start := tzr.Pos - 2
	for {
		if tzr.Cur() == '*' {
			tzr.Skip(1)
			if tzr.Cur() == '/' {
				tzr.Skip(1)
				break
			}
			continue
		}
		if tzr.Cur() == tokenizer.EofChar {
			return LEX_ERROR, tzr.buf[start:tzr.Pos]
		}
		tzr.Skip(1)
	}
	return COMMENT, tzr.buf[start:tzr.Pos]
}

// scanPSQLSpecificComment scans a PSQL comment pragma, which always starts with '//*`
func (tzr *PsqlTokenizer) scanPSQLSpecificComment() (int, string) {
	start := tzr.Pos - 3
	for {
		if tzr.Cur() == '*' {
			tzr.Skip(1)
			if tzr.Cur() == '/' {
				tzr.Skip(1)
				break
			}
			continue
		}
		if tzr.Cur() == tokenizer.EofChar {
			return LEX_ERROR, tzr.buf[start:tzr.Pos]
		}
		tzr.Skip(1)
	}

	commentVersion, sql := ExtractPsqlComment(tzr.buf[start:tzr.Pos])

	if "1" >= commentVersion {
		// Only add the special comment to the tokenizer if the version of PSQL is higher or equal to the comment version
		tzr.specialComment = NewPsqlStringTokenizer(sql)
	}

	return tzr.Scan()
}

func (tzr *PsqlTokenizer) Cur() rune {
	return tzr.Peek(0)
}

func (tzr *PsqlTokenizer) Skip(dist int) {
	tzr.Pos += dist
}

func (tzr *PsqlTokenizer) Peek(dist int) rune {
	if tzr.Pos+dist >= len(tzr.buf) {
		return tokenizer.EofChar
	}
	return rune(tzr.buf[tzr.Pos+dist])
}

// Reset clears any internal state.
func (tzr *PsqlTokenizer) Reset() {
	tzr.ParseTree = nil
	tzr.partialDDL = nil
	tzr.specialComment = nil
	tzr.posVarIndex = 0
	tzr.nesting = 0
	tzr.SkipToEnd = false
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '$'
}

func isCarat(ch rune) bool {
	return ch == '.' || ch == '\'' || ch == '"' || ch == '`'
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch) - '0'
	case 'a' <= ch && ch <= 'f':
		return int(ch) - 'a' + 10
	case 'A' <= ch && ch <= 'F':
		return int(ch) - 'A' + 10
	}
	return 16 // larger than any legal digit val
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

// ExtractPsqlComment extracts the version and SQL from a comment-only query
// such as /*!50708 sql here */
func ExtractPsqlComment(sql string) (string, string) {
	sql = sql[3 : len(sql)-2]

	digitCount := 0
	endOfVersionIndex := strings.IndexFunc(sql, func(c rune) bool {
		digitCount++
		return !unicode.IsDigit(c) || digitCount == 6
	})
	if endOfVersionIndex < 0 {
		return "", ""
	}
	if endOfVersionIndex < 5 {
		endOfVersionIndex = 0
	}
	version := sql[0:endOfVersionIndex]
	innerSQL := strings.TrimFunc(sql[endOfVersionIndex:], unicode.IsSpace)

	return version, innerSQL
}
