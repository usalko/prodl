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

	Pos int
	buf string
}

// SetSkipSpecialComments implements tokenizer.Tokenizer.
func (tkn *PsqlTokenizer) SetSkipSpecialComments(skip bool) {
	tkn.SkipSpecialComments = skip
}

// GetBindVars implements tokenizer.Tokenizer.
func (tkn *PsqlTokenizer) GetBindVars() ast.BindVars {
	return tkn.BindVars
}

// GetLastError implements tokenizer.Tokenizer.
func (tkn *PsqlTokenizer) GetLastError() error {
	return tkn.LastError
}

// GetPos implements tokenizer.Tokenizer.
func (tkn *PsqlTokenizer) GetPos() int {
	return tkn.Pos
}

// SetMulti implements tokenizer.Tokenizer.
func (tkn *PsqlTokenizer) SetMulti(multi bool) {
	tkn.multi = multi
}

// GetIdToken implements sql_parser.Tokenizer.
func (tkn *PsqlTokenizer) GetIdToken() int {
	return ID
}

// GetKeywordString implements sql_parser.Tokenizer.
func (tkn *PsqlTokenizer) GetKeywordString(token int) string {
	return KeywordString(token)
}

// GetParseTree implements sql_parser.Tokenizer.
func (tkn *PsqlTokenizer) GetParseTree() ast.Statement {
	return tkn.ParseTree
}

// GetPartialDDL implements sql_parser.Tokenizer.
func (tkn *PsqlTokenizer) GetPartialDDL() ast.Statement {
	return tkn.partialDDL
}

// BindVar implements sql_parser.Tokenizer.
func (tkn *PsqlTokenizer) BindVar(bvar string, value struct{}) {
	tkn.BindVars[bvar] = value
}

// DecNesting implements sql_parser.Tokenizer.
func (tkn *PsqlTokenizer) DecNesting() {
	tkn.nesting--
}

// GetNesting implements sql_parser.Tokenizer.
func (tkn *PsqlTokenizer) GetNesting() int {
	return tkn.nesting
}

// IncNesting implements sql_parser.Tokenizer.
func (tkn *PsqlTokenizer) IncNesting() {
	tkn.nesting++
}

// SetAllowComments implements sql_parser.Tokenizer.
func (tkn *PsqlTokenizer) SetAllowComments(allow bool) {
	tkn.AllowComments = allow
}

// SetParseTree implements sql_parser.Tokenizer.
func (tkn *PsqlTokenizer) SetParseTree(stmt ast.Statement) {
	tkn.ParseTree = stmt
}

// SetPartialDDL implements sql_parser.Tokenizer.
func (tkn *PsqlTokenizer) SetPartialDDL(node ast.Statement) {
	tkn.partialDDL = node
}

// SetSkipToEnd implements sql_parser.Tokenizer.
func (tkn *PsqlTokenizer) SetSkipToEnd(skip bool) {
	tkn.SkipToEnd = skip
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
		buf:      sql,
		BindVars: make(map[string]struct{}),
	}
}

// Lex returns the next token form the Tokenizer.
// This function is used by go yacc.
func (tkn *PsqlTokenizer) Lex(lval *psqSymType) int {
	if tkn.SkipToEnd {
		return tkn.skipStatement()
	}

	typ, val := tkn.Scan()
	for typ == COMMENT {
		if tkn.AllowComments || val == "COMMENT" {
			break
		}
		typ, val = tkn.Scan()
	}
	if typ == 0 || typ == ';' || typ == LEX_ERROR {
		// If encounter end of statement or invalid token,
		// we should not accept partially parsed DDLs. They
		// should instead result in parser errors. See the
		// Parse function to see how this is handled.
		tkn.partialDDL = nil
	}
	lval.str = val
	tkn.lastToken = val
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
func (tkn *PsqlTokenizer) Error(err string) {
	tkn.LastError = PositionedErr{Err: err, Pos: tkn.Pos + 1, Near: tkn.lastToken}

	// Try and re-sync to the next statement
	tkn.skipStatement()
}

// Scan scans the tokenizer for the next token and returns
// the token type and an optional value.
func (tkn *PsqlTokenizer) Scan() (int, string) {
	if tkn.specialComment != nil {
		// Enter specialComment scan mode.
		// for scanning such kind of comment: /*! PSQL-specific code */
		specialComment := tkn.specialComment
		tok, val := specialComment.Scan()
		if tok != 0 {
			// return the specialComment scan result as the result
			return tok, val
		}
		// leave specialComment scan mode after all stream consumed.
		tkn.specialComment = nil
	}

	tkn.SkipBlank()
	switch ch := tkn.Cur(); {
	case ch == '@':
		tokenID := AT_ID
		tkn.Skip(1)
		if tkn.Cur() == '@' {
			tokenID = AT_AT_ID
			tkn.Skip(1)
		}
		var tID int
		var tBytes string
		if tkn.Cur() == '`' {
			tkn.Skip(1)
			tID, tBytes = tkn.scanLiteralIdentifier()
		} else if tkn.Cur() == tokenizer.EofChar {
			return LEX_ERROR, ""
		} else {
			tID, tBytes = tkn.scanIdentifier(true)
		}
		if tID == LEX_ERROR {
			return tID, ""
		}
		return tokenID, tBytes
	case isLetter(ch):
		if ch == 'X' || ch == 'x' {
			if tkn.Peek(1) == '\'' {
				tkn.Skip(2)
				return tkn.scanHex()
			}
		}
		if ch == 'B' || ch == 'b' {
			if tkn.Peek(1) == '\'' {
				tkn.Skip(2)
				return tkn.scanBitLiteral()
			}
		}
		// N\'literal' is used to create a string in the national character set
		if ch == 'N' || ch == 'n' {
			nxt := tkn.Peek(1)
			if nxt == '\'' || nxt == '"' {
				tkn.Skip(2)
				return tkn.scanString(nxt, NCHAR_STRING)
			}
		}
		return tkn.scanIdentifier(false)
	case isDigit(ch):
		return tkn.scanNumber()
	case ch == ':':
		return tkn.scanBindVar()
	case ch == ';':
		if tkn.multi {
			// In multi mode, ';' is treated as EOF. So, we don't advance.
			// Repeated calls to Scan will keep returning 0 until ParseNext
			// forces the advance.
			return 0, ""
		}
		tkn.Skip(1)
		return ';', ""
	case ch == tokenizer.EofChar:
		return 0, ""
	default:
		if ch == '.' && isDigit(tkn.Peek(1)) {
			return tkn.scanNumber()
		}

		tkn.Skip(1)
		switch ch {
		case '=', ',', '(', ')', '+', '*', '%', '^', '~':
			return int(ch), ""
		case '&':
			if tkn.Cur() == '&' {
				tkn.Skip(1)
				return AND, ""
			}
			return int(ch), ""
		case '|':
			if tkn.Cur() == '|' {
				tkn.Skip(1)
				return OR, ""
			}
			return int(ch), ""
		case '?':
			tkn.posVarIndex++
			buf := make([]byte, 0, 8)
			buf = append(buf, ":v"...)
			buf = strconv.AppendInt(buf, int64(tkn.posVarIndex), 10)
			return VALUE_ARG, string(buf)
		case '.':
			return int(ch), ""
		case '/':
			switch tkn.Cur() {
			case '/':
				tkn.Skip(1)
				return tkn.scanCommentType1(2)
			case '*':
				tkn.Skip(1)
				if tkn.Cur() == '!' && !tkn.SkipSpecialComments {
					tkn.Skip(1)
					return tkn.scanPSQLSpecificComment()
				}
				return tkn.scanCommentType2()
			default:
				return int(ch), ""
			}
		case '#':
			return tkn.scanCommentType1(1)
		case '-':
			switch tkn.Cur() {
			case '-':
				nextChar := tkn.Peek(1)
				if nextChar == ' ' || nextChar == '\n' || nextChar == '\t' || nextChar == '\r' || nextChar == tokenizer.EofChar {
					tkn.Skip(1)
					return tkn.scanCommentType1(2)
				}
			case '>':
				tkn.Skip(1)
				if tkn.Cur() == '>' {
					tkn.Skip(1)
					return JSON_UNQUOTE_EXTRACT_OP, ""
				}
				return JSON_EXTRACT_OP, ""
			}
			return int(ch), ""
		case '<':
			switch tkn.Cur() {
			case '>':
				tkn.Skip(1)
				return NE, ""
			case '<':
				tkn.Skip(1)
				return SHIFT_LEFT, ""
			case '=':
				tkn.Skip(1)
				switch tkn.Cur() {
				case '>':
					tkn.Skip(1)
					return NULL_SAFE_EQUAL, ""
				default:
					return LE, ""
				}
			default:
				return int(ch), ""
			}
		case '>':
			switch tkn.Cur() {
			case '=':
				tkn.Skip(1)
				return GE, ""
			case '>':
				tkn.Skip(1)
				return SHIFT_RIGHT, ""
			default:
				return int(ch), ""
			}
		case '!':
			if tkn.Cur() == '=' {
				tkn.Skip(1)
				return NE, ""
			}
			return int(ch), ""
		case '\'', '"':
			return tkn.scanString(ch, STRING)
		case '`':
			return tkn.scanLiteralIdentifier()
		default:
			return LEX_ERROR, string(byte(ch))
		}
	}
}

// skipStatement scans until end of statement.
func (tkn *PsqlTokenizer) skipStatement() int {
	tkn.SkipToEnd = false
	for {
		typ, _ := tkn.Scan()
		if typ == 0 || typ == ';' || typ == LEX_ERROR {
			return typ
		}
	}
}

// SkipBlank skips the cursor while it finds whitespace
func (tkn *PsqlTokenizer) SkipBlank() {
	ch := tkn.Cur()
	for ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' {
		tkn.Skip(1)
		ch = tkn.Cur()
	}
}

// scanIdentifier scans a language keyword or @-encased variable
func (tkn *PsqlTokenizer) scanIdentifier(isVariable bool) (int, string) {
	start := tkn.Pos
	tkn.Skip(1)

	for {
		ch := tkn.Cur()
		if !isLetter(ch) && !isDigit(ch) && !(isVariable && isCarat(ch)) {
			break
		}
		tkn.Skip(1)
	}
	keywordName := tkn.buf[start:tkn.Pos]
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
func (tkn *PsqlTokenizer) scanHex() (int, string) {
	start := tkn.Pos
	tkn.scanMantissa(16)
	hex := tkn.buf[start:tkn.Pos]
	if tkn.Cur() != '\'' {
		return LEX_ERROR, hex
	}
	tkn.Skip(1)
	if len(hex)%2 != 0 {
		return LEX_ERROR, hex
	}
	return HEX, hex
}

// scanBitLiteral scans a binary numeric literal; assumes b' or B' has already been scanned
func (tkn *PsqlTokenizer) scanBitLiteral() (int, string) {
	start := tkn.Pos
	tkn.scanMantissa(2)
	bit := tkn.buf[start:tkn.Pos]
	if tkn.Cur() != '\'' {
		return LEX_ERROR, bit
	}
	tkn.Skip(1)
	return BIT_LITERAL, bit
}

// scanLiteralIdentifierSlow scans an identifier surrounded by backticks which may
// contain escape sequences instead of it. This method is only called from
// scanLiteralIdentifier once the first escape sequence is found in the identifier.
// The provided `buf` contains the contents of the identifier that have been scanned
// so far.
func (tkn *PsqlTokenizer) scanLiteralIdentifierSlow(buf *strings.Builder) (int, string) {
	backTickSeen := true
	for {
		if backTickSeen {
			if tkn.Cur() != '`' {
				break
			}
			backTickSeen = false
			buf.WriteByte('`')
			tkn.Skip(1)
			continue
		}
		// The previous char was not a backtick.
		switch tkn.Cur() {
		case '`':
			backTickSeen = true
		case tokenizer.EofChar:
			// Premature EOF.
			return LEX_ERROR, buf.String()
		default:
			buf.WriteByte(byte(tkn.Cur()))
			// keep scanning
		}
		tkn.Skip(1)
	}
	return ID, buf.String()
}

// scanLiteralIdentifier scans an identifier enclosed by backticks. If the identifier
// is a simple literal, it'll be returned as a slice of the input buffer. If the identifier
// contains escape sequences, this function will fall back to scanLiteralIdentifierSlow
func (tkn *PsqlTokenizer) scanLiteralIdentifier() (int, string) {
	start := tkn.Pos
	for {
		switch tkn.Cur() {
		case '`':
			if tkn.Peek(1) != '`' {
				if tkn.Pos == start {
					return LEX_ERROR, ""
				}
				tkn.Skip(1)
				return ID, tkn.buf[start : tkn.Pos-1]
			}

			var buf strings.Builder
			buf.WriteString(tkn.buf[start:tkn.Pos])
			tkn.Skip(1)
			return tkn.scanLiteralIdentifierSlow(&buf)
		case tokenizer.EofChar:
			// Premature EOF.
			return LEX_ERROR, tkn.buf[start:tkn.Pos]
		default:
			tkn.Skip(1)
		}
	}
}

// scanBindVar scans a bind variable; assumes a ':' has been scanned right before
func (tkn *PsqlTokenizer) scanBindVar() (int, string) {
	start := tkn.Pos
	token := VALUE_ARG

	tkn.Skip(1)
	if tkn.Cur() == ':' {
		token = LIST_ARG
		tkn.Skip(1)
	}
	if !isLetter(tkn.Cur()) {
		return LEX_ERROR, tkn.buf[start:tkn.Pos]
	}
	for {
		ch := tkn.Cur()
		if !isLetter(ch) && !isDigit(ch) && ch != '.' {
			break
		}
		tkn.Skip(1)
	}
	return token, tkn.buf[start:tkn.Pos]
}

// scanMantissa scans a sequence of numeric characters with the same base.
// This is a helper function only called from the numeric scanners
func (tkn *PsqlTokenizer) scanMantissa(base int) {
	for digitVal(tkn.Cur()) < base {
		tkn.Skip(1)
	}
}

// scanNumber scans any SQL numeric literal, either floating point or integer
func (tkn *PsqlTokenizer) scanNumber() (int, string) {
	start := tkn.Pos
	token := INTEGRAL

	if tkn.Cur() == '.' {
		token = DECIMAL
		tkn.Skip(1)
		tkn.scanMantissa(10)
		goto exponent
	}

	// 0x construct.
	if tkn.Cur() == '0' {
		tkn.Skip(1)
		if tkn.Cur() == 'x' || tkn.Cur() == 'X' {
			token = HEXNUM
			tkn.Skip(1)
			tkn.scanMantissa(16)
			goto exit
		}
	}

	tkn.scanMantissa(10)

	if tkn.Cur() == '.' {
		token = DECIMAL
		tkn.Skip(1)
		tkn.scanMantissa(10)
	}

exponent:
	if tkn.Cur() == 'e' || tkn.Cur() == 'E' {
		token = FLOAT
		tkn.Skip(1)
		if tkn.Cur() == '+' || tkn.Cur() == '-' {
			tkn.Skip(1)
		}
		tkn.scanMantissa(10)
	}

exit:
	if isLetter(tkn.Cur()) {
		// A letter cannot immediately follow a float number.
		if token == FLOAT || token == DECIMAL {
			return LEX_ERROR, tkn.buf[start:tkn.Pos]
		}
		// A letter seen after a few numbers means that we should parse this
		// as an identifier and not a number.
		for {
			ch := tkn.Cur()
			if !isLetter(ch) && !isDigit(ch) {
				break
			}
			tkn.Skip(1)
		}
		return ID, tkn.buf[start:tkn.Pos]
	}

	return token, tkn.buf[start:tkn.Pos]
}

// scanString scans a string surrounded by the given `delim`, which can be
// either single or double quotes. Assumes that the given delimiter has just
// been scanned. If the skin contains any escape sequences, this function
// will fall back to scanStringSlow
func (tkn *PsqlTokenizer) scanString(delim uint16, typ int) (int, string) {
	start := tkn.Pos

	for {
		switch tkn.Cur() {
		case delim:
			if tkn.Peek(1) != delim {
				tkn.Skip(1)
				return typ, tkn.buf[start : tkn.Pos-1]
			}
			fallthrough

		case '\\':
			var buffer strings.Builder
			buffer.WriteString(tkn.buf[start:tkn.Pos])
			return tkn.scanStringSlow(&buffer, delim, typ)

		case tokenizer.EofChar:
			return LEX_ERROR, tkn.buf[start:tkn.Pos]
		}

		tkn.Skip(1)
	}
}

// scanString scans a string surrounded by the given `delim` and containing escape
// sequencse. The given `buffer` contains the contents of the string that have
// been scanned so far.
func (tkn *PsqlTokenizer) scanStringSlow(buffer *strings.Builder, delim uint16, typ int) (int, string) {
	for {
		ch := tkn.Cur()
		if ch == tokenizer.EofChar {
			// Unterminated string.
			return LEX_ERROR, buffer.String()
		}

		if ch != delim && ch != '\\' {
			// Scan ahead to the next interesting character.
			start := tkn.Pos
			for ; tkn.Pos < len(tkn.buf); tkn.Pos++ {
				ch = uint16(tkn.buf[tkn.Pos])
				if ch == delim || ch == '\\' {
					break
				}
			}

			buffer.WriteString(tkn.buf[start:tkn.Pos])
			if tkn.Pos >= len(tkn.buf) {
				// Reached the end of the buffer without finding a delim or
				// escape character.
				tkn.Skip(1)
				continue
			}
		}
		tkn.Skip(1) // Read one past the delim or escape character.

		if ch == '\\' {
			if tkn.Cur() == tokenizer.EofChar {
				// String terminates mid escape character.
				return LEX_ERROR, buffer.String()
			}
			if decodedChar := sql_types.SQLDecodeMap[byte(tkn.Cur())]; decodedChar == sql_types.DontEscape {
				ch = tkn.Cur()
			} else {
				ch = uint16(decodedChar)
			}
		} else if ch == delim && tkn.Cur() != delim {
			// Correctly terminated string, which is not a double delim.
			break
		}

		buffer.WriteByte(byte(ch))
		tkn.Skip(1)
	}

	return typ, buffer.String()
}

// scanCommentType1 scans a SQL line-comment, which is applied until the end
// of the line. The given prefix length varies based on whether the comment
// is started with '//', '--' or '#'.
func (tkn *PsqlTokenizer) scanCommentType1(prefixLen int) (int, string) {
	start := tkn.Pos - prefixLen
	for tkn.Cur() != tokenizer.EofChar {
		if tkn.Cur() == '\n' {
			tkn.Skip(1)
			break
		}
		tkn.Skip(1)
	}
	return COMMENT, tkn.buf[start:tkn.Pos]
}

// scanCommentType2 scans a '/*' delimited comment; assumes the opening
// prefix has already been scanned
func (tkn *PsqlTokenizer) scanCommentType2() (int, string) {
	start := tkn.Pos - 2
	for {
		if tkn.Cur() == '*' {
			tkn.Skip(1)
			if tkn.Cur() == '/' {
				tkn.Skip(1)
				break
			}
			continue
		}
		if tkn.Cur() == tokenizer.EofChar {
			return LEX_ERROR, tkn.buf[start:tkn.Pos]
		}
		tkn.Skip(1)
	}
	return COMMENT, tkn.buf[start:tkn.Pos]
}

// scanPSQLSpecificComment scans a PSQL comment pragma, which always starts with '//*`
func (tkn *PsqlTokenizer) scanPSQLSpecificComment() (int, string) {
	start := tkn.Pos - 3
	for {
		if tkn.Cur() == '*' {
			tkn.Skip(1)
			if tkn.Cur() == '/' {
				tkn.Skip(1)
				break
			}
			continue
		}
		if tkn.Cur() == tokenizer.EofChar {
			return LEX_ERROR, tkn.buf[start:tkn.Pos]
		}
		tkn.Skip(1)
	}

	commentVersion, sql := ExtractPsqlComment(tkn.buf[start:tkn.Pos])

	if "1" >= commentVersion {
		// Only add the special comment to the tokenizer if the version of PSQL is higher or equal to the comment version
		tkn.specialComment = NewPsqlStringTokenizer(sql)
	}

	return tkn.Scan()
}

func (tkn *PsqlTokenizer) Cur() uint16 {
	return tkn.Peek(0)
}

func (tkn *PsqlTokenizer) Skip(dist int) {
	tkn.Pos += dist
}

func (tkn *PsqlTokenizer) Peek(dist int) uint16 {
	if tkn.Pos+dist >= len(tkn.buf) {
		return tokenizer.EofChar
	}
	return uint16(tkn.buf[tkn.Pos+dist])
}

// Reset clears any internal state.
func (tkn *PsqlTokenizer) Reset() {
	tkn.ParseTree = nil
	tkn.partialDDL = nil
	tkn.specialComment = nil
	tkn.posVarIndex = 0
	tkn.nesting = 0
	tkn.SkipToEnd = false
}

func isLetter(ch uint16) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '$'
}

func isCarat(ch uint16) bool {
	return ch == '.' || ch == '\'' || ch == '"' || ch == '`'
}

func digitVal(ch uint16) int {
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

func isDigit(ch uint16) bool {
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
