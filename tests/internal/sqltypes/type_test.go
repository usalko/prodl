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

package sqltypes

import (
	"strings"
	"testing"

	"github.com/usalko/sent/internal/sqltypes"
)

func TestTypeValues(t *testing.T) {
	testcases := []struct {
		defined  sqltypes.Type
		expected int
	}{{
		defined:  sqltypes.Null,
		expected: 0,
	}, {
		defined:  sqltypes.Int8,
		expected: 1 | sqltypes.FlagIsIntegral,
	}, {
		defined:  sqltypes.Uint8,
		expected: 2 | sqltypes.FlagIsIntegral | sqltypes.FlagIsUnsigned,
	}, {
		defined:  sqltypes.Int16,
		expected: 3 | sqltypes.FlagIsIntegral,
	}, {
		defined:  sqltypes.Uint16,
		expected: 4 | sqltypes.FlagIsIntegral | sqltypes.FlagIsUnsigned,
	}, {
		defined:  sqltypes.Int24,
		expected: 5 | sqltypes.FlagIsIntegral,
	}, {
		defined:  sqltypes.Uint24,
		expected: 6 | sqltypes.FlagIsIntegral | sqltypes.FlagIsUnsigned,
	}, {
		defined:  sqltypes.Int32,
		expected: 7 | sqltypes.FlagIsIntegral,
	}, {
		defined:  sqltypes.Uint32,
		expected: 8 | sqltypes.FlagIsIntegral | sqltypes.FlagIsUnsigned,
	}, {
		defined:  sqltypes.Int64,
		expected: 9 | sqltypes.FlagIsIntegral,
	}, {
		defined:  sqltypes.Uint64,
		expected: 10 | sqltypes.FlagIsIntegral | sqltypes.FlagIsUnsigned,
	}, {
		defined:  sqltypes.Float32,
		expected: 11 | sqltypes.FlagIsFloat,
	}, {
		defined:  sqltypes.Float64,
		expected: 12 | sqltypes.FlagIsFloat,
	}, {
		defined:  sqltypes.Timestamp,
		expected: 13 | sqltypes.FlagIsQuoted,
	}, {
		defined:  sqltypes.Date,
		expected: 14 | sqltypes.FlagIsQuoted,
	}, {
		defined:  sqltypes.Time,
		expected: 15 | sqltypes.FlagIsQuoted,
	}, {
		defined:  sqltypes.Datetime,
		expected: 16 | sqltypes.FlagIsQuoted,
	}, {
		defined:  sqltypes.Year,
		expected: 17 | sqltypes.FlagIsIntegral | sqltypes.FlagIsUnsigned,
	}, {
		defined:  sqltypes.Decimal,
		expected: 18,
	}, {
		defined:  sqltypes.Text,
		expected: 19 | sqltypes.FlagIsQuoted | sqltypes.FlagIsText,
	}, {
		defined:  sqltypes.Blob,
		expected: 20 | sqltypes.FlagIsQuoted | sqltypes.FlagIsBinary,
	}, {
		defined:  sqltypes.VarChar,
		expected: 21 | sqltypes.FlagIsQuoted | sqltypes.FlagIsText,
	}, {
		defined:  sqltypes.VarBinary,
		expected: 22 | sqltypes.FlagIsQuoted | sqltypes.FlagIsBinary,
	}, {
		defined:  sqltypes.Char,
		expected: 23 | sqltypes.FlagIsQuoted | sqltypes.FlagIsText,
	}, {
		defined:  sqltypes.Binary,
		expected: 24 | sqltypes.FlagIsQuoted | sqltypes.FlagIsBinary,
	}, {
		defined:  sqltypes.Bit,
		expected: 25 | sqltypes.FlagIsQuoted,
	}, {
		defined:  sqltypes.Enum,
		expected: 26 | sqltypes.FlagIsQuoted,
	}, {
		defined:  sqltypes.Set,
		expected: 27 | sqltypes.FlagIsQuoted,
	}, {
		defined:  sqltypes.Geometry,
		expected: 29 | sqltypes.FlagIsQuoted,
	}, {
		defined:  sqltypes.TypeJSON,
		expected: 30 | sqltypes.FlagIsQuoted,
	}, {
		defined:  sqltypes.Expression,
		expected: 31,
	}, {
		defined:  sqltypes.HexNum,
		expected: 32 | sqltypes.FlagIsText,
	}, {
		defined:  sqltypes.HexVal,
		expected: 33 | sqltypes.FlagIsText,
	}}
	for _, tcase := range testcases {
		if int(tcase.defined) != tcase.expected {
			t.Errorf("Type %s: %d, want: %d", tcase.defined, int(tcase.defined), tcase.expected)
		}
	}
}

// TestCategory verifies that the type categorizations
// are non-overlapping and complete.
func TestCategory(t *testing.T) {
	alltypes := []sqltypes.Type{
		sqltypes.Null,
		sqltypes.Int8,
		sqltypes.Uint8,
		sqltypes.Int16,
		sqltypes.Uint16,
		sqltypes.Int24,
		sqltypes.Uint24,
		sqltypes.Int32,
		sqltypes.Uint32,
		sqltypes.Int64,
		sqltypes.Uint64,
		sqltypes.Float32,
		sqltypes.Float64,
		sqltypes.Timestamp,
		sqltypes.Date,
		sqltypes.Time,
		sqltypes.Datetime,
		sqltypes.Year,
		sqltypes.Decimal,
		sqltypes.Text,
		sqltypes.Blob,
		sqltypes.VarChar,
		sqltypes.VarBinary,
		sqltypes.Char,
		sqltypes.Binary,
		sqltypes.Bit,
		sqltypes.Enum,
		sqltypes.Set,
		sqltypes.Geometry,
		sqltypes.TypeJSON,
		sqltypes.Expression,
		sqltypes.HexNum,
		sqltypes.HexVal,
	}
	for _, typ := range alltypes {
		matched := false
		if sqltypes.IsSigned(typ) {
			if !sqltypes.IsIntegral(typ) {
				t.Errorf("Signed type %v is not an integral", typ)
			}
			matched = true
		}
		if sqltypes.IsUnsigned(typ) {
			if !sqltypes.IsIntegral(typ) {
				t.Errorf("Unsigned type %v is not an integral", typ)
			}
			if matched {
				t.Errorf("%v matched more than one category", typ)
			}
			matched = true
		}
		if sqltypes.IsFloat(typ) {
			if matched {
				t.Errorf("%v matched more than one category", typ)
			}
			matched = true
		}
		if sqltypes.IsQuoted(typ) {
			if matched {
				t.Errorf("%v matched more than one category", typ)
			}
			matched = true
		}
		if typ == sqltypes.Null || typ == sqltypes.Decimal || typ == sqltypes.Expression || typ == sqltypes.Bit || typ == sqltypes.HexNum || typ == sqltypes.HexVal {
			if matched {
				t.Errorf("%v matched more than one category", typ)
			}
			matched = true
		}
		if !matched {
			t.Errorf("%v matched no category", typ)
		}
	}
}

func TestIsFunctions(t *testing.T) {
	if sqltypes.IsIntegral(sqltypes.Null) {
		t.Error("Null: IsIntegral, must be false")
	}
	if !sqltypes.IsIntegral(sqltypes.Int64) {
		t.Error("Int64: !IsIntegral, must be true")
	}
	if sqltypes.IsSigned(sqltypes.Uint64) {
		t.Error("Uint64: IsSigned, must be false")
	}
	if !sqltypes.IsSigned(sqltypes.Int64) {
		t.Error("Int64: !IsSigned, must be true")
	}
	if sqltypes.IsUnsigned(sqltypes.Int64) {
		t.Error("Int64: IsUnsigned, must be false")
	}
	if !sqltypes.IsUnsigned(sqltypes.Uint64) {
		t.Error("Uint64: !IsUnsigned, must be true")
	}
	if sqltypes.IsFloat(sqltypes.Int64) {
		t.Error("Int64: IsFloat, must be false")
	}
	if !sqltypes.IsFloat(sqltypes.Float64) {
		t.Error("Uint64: !IsFloat, must be true")
	}
	if sqltypes.IsQuoted(sqltypes.Int64) {
		t.Error("Int64: IsQuoted, must be false")
	}
	if !sqltypes.IsQuoted(sqltypes.Binary) {
		t.Error("Binary: !IsQuoted, must be true")
	}
	if sqltypes.IsText(sqltypes.Int64) {
		t.Error("Int64: IsText, must be false")
	}
	if !sqltypes.IsText(sqltypes.Char) {
		t.Error("Char: !IsText, must be true")
	}
	if sqltypes.IsBinary(sqltypes.Int64) {
		t.Error("Int64: IsBinary, must be false")
	}
	if !sqltypes.IsBinary(sqltypes.Binary) {
		t.Error("Char: !IsBinary, must be true")
	}
	if !sqltypes.IsNumber(sqltypes.Int64) {
		t.Error("Int64: !isNumber, must be true")
	}
}

func TestTypeToMySQL(t *testing.T) {
	v, f := sqltypes.TypeToMySQL(sqltypes.Bit)
	if v != 16 {
		t.Errorf("Bit: %d, want 16", v)
	}
	if f != sqltypes.MysqlUnsigned {
		t.Errorf("Bit flag: %x, want %x", f, sqltypes.MysqlUnsigned)
	}
	v, f = sqltypes.TypeToMySQL(sqltypes.Date)
	if v != 10 {
		t.Errorf("Bit: %d, want 10", v)
	}
	if f != sqltypes.MysqlBinary {
		t.Errorf("Bit flag: %x, want %x", f, sqltypes.MysqlBinary)
	}
}

func TestMySQLToType(t *testing.T) {
	testcases := []struct {
		intype  int64
		inflags int64
		outtype sqltypes.Type
	}{{
		intype:  1,
		outtype: sqltypes.Int8,
	}, {
		intype:  1,
		inflags: sqltypes.MysqlUnsigned,
		outtype: sqltypes.Uint8,
	}, {
		intype:  2,
		outtype: sqltypes.Int16,
	}, {
		intype:  2,
		inflags: sqltypes.MysqlUnsigned,
		outtype: sqltypes.Uint16,
	}, {
		intype:  3,
		outtype: sqltypes.Int32,
	}, {
		intype:  3,
		inflags: sqltypes.MysqlUnsigned,
		outtype: sqltypes.Uint32,
	}, {
		intype:  4,
		outtype: sqltypes.Float32,
	}, {
		intype:  5,
		outtype: sqltypes.Float64,
	}, {
		intype:  6,
		outtype: sqltypes.Null,
	}, {
		intype:  7,
		outtype: sqltypes.Timestamp,
	}, {
		intype:  8,
		outtype: sqltypes.Int64,
	}, {
		intype:  8,
		inflags: sqltypes.MysqlUnsigned,
		outtype: sqltypes.Uint64,
	}, {
		intype:  9,
		outtype: sqltypes.Int24,
	}, {
		intype:  9,
		inflags: sqltypes.MysqlUnsigned,
		outtype: sqltypes.Uint24,
	}, {
		intype:  10,
		outtype: sqltypes.Date,
	}, {
		intype:  11,
		outtype: sqltypes.Time,
	}, {
		intype:  12,
		outtype: sqltypes.Datetime,
	}, {
		intype:  13,
		outtype: sqltypes.Year,
	}, {
		intype:  16,
		outtype: sqltypes.Bit,
	}, {
		intype:  245,
		outtype: sqltypes.TypeJSON,
	}, {
		intype:  246,
		outtype: sqltypes.Decimal,
	}, {
		intype:  249,
		outtype: sqltypes.Text,
	}, {
		intype:  250,
		outtype: sqltypes.Text,
	}, {
		intype:  251,
		outtype: sqltypes.Text,
	}, {
		intype:  252,
		outtype: sqltypes.Text,
	}, {
		intype:  252,
		inflags: sqltypes.MysqlBinary,
		outtype: sqltypes.Blob,
	}, {
		intype:  253,
		outtype: sqltypes.VarChar,
	}, {
		intype:  253,
		inflags: sqltypes.MysqlBinary,
		outtype: sqltypes.VarBinary,
	}, {
		intype:  254,
		outtype: sqltypes.Char,
	}, {
		intype:  254,
		inflags: sqltypes.MysqlBinary,
		outtype: sqltypes.Binary,
	}, {
		intype:  254,
		inflags: sqltypes.MysqlEnum,
		outtype: sqltypes.Enum,
	}, {
		intype:  254,
		inflags: sqltypes.MysqlSet,
		outtype: sqltypes.Set,
	}, {
		intype:  255,
		outtype: sqltypes.Geometry,
	}, {
		// Binary flag must be ignored.
		intype:  8,
		inflags: sqltypes.MysqlUnsigned | sqltypes.MysqlBinary,
		outtype: sqltypes.Uint64,
	}, {
		// Unsigned flag must be ignored
		intype:  252,
		inflags: sqltypes.MysqlUnsigned | sqltypes.MysqlBinary,
		outtype: sqltypes.Blob,
	}}
	for _, tcase := range testcases {
		got, err := sqltypes.MySQLToType(tcase.intype, tcase.inflags)
		if err != nil {
			t.Error(err)
		}
		if got != tcase.outtype {
			t.Errorf("MySQLToType(%d, %x): %v, want %v", tcase.intype, tcase.inflags, got, tcase.outtype)
		}
	}
}

func TestTypeError(t *testing.T) {
	_, err := sqltypes.MySQLToType(50, 0)
	want := "unsupported type: 50"
	if err == nil || err.Error() != want {
		t.Errorf("MySQLToType: %v, want %s", err, want)
	}
}

func TestTypeEquivalenceCheck(t *testing.T) {
	if !sqltypes.AreTypesEquivalent(sqltypes.Int16, sqltypes.Int16) {
		t.Errorf("Int16 and Int16 are same types.")
	}
	if sqltypes.AreTypesEquivalent(sqltypes.Int16, sqltypes.Int24) {
		t.Errorf("Int16 and Int24 are not same types.")
	}
	if !sqltypes.AreTypesEquivalent(sqltypes.VarChar, sqltypes.VarBinary) {
		t.Errorf("VarChar in binlog and VarBinary in schema are equivalent types.")
	}
	if sqltypes.AreTypesEquivalent(sqltypes.VarBinary, sqltypes.VarChar) {
		t.Errorf("VarBinary in binlog and VarChar in schema are not equivalent types.")
	}
	if !sqltypes.AreTypesEquivalent(sqltypes.Int16, sqltypes.Uint16) {
		t.Errorf("Int16 in binlog and Uint16 in schema are equivalent types.")
	}
	if sqltypes.AreTypesEquivalent(sqltypes.Uint16, sqltypes.Int16) {
		t.Errorf("Uint16 in binlog and Int16 in schema are not equivalent types.")
	}
}

func TestPrintTypeChecks(t *testing.T) {
	var funcs = []struct {
		name string
		f    func(p sqltypes.Type) bool
	}{
		{"IsSigned", sqltypes.IsSigned},
		{"IsFloat", sqltypes.IsFloat},
		{"IsUnsigned", sqltypes.IsUnsigned},
		{"IsIntegral", sqltypes.IsIntegral},
		{"IsText", sqltypes.IsText},
		{"IsNumber", sqltypes.IsNumber},
		{"IsQuoted", sqltypes.IsQuoted},
		{"IsBinary", sqltypes.IsBinary},
		{"IsDate", sqltypes.IsDate},
		{"IsNull", sqltypes.IsNull},
	}
	var types = []sqltypes.Type{
		sqltypes.Null,
		sqltypes.Int8,
		sqltypes.Uint8,
		sqltypes.Int16,
		sqltypes.Uint16,
		sqltypes.Int24,
		sqltypes.Uint24,
		sqltypes.Int32,
		sqltypes.Uint32,
		sqltypes.Int64,
		sqltypes.Uint64,
		sqltypes.Float32,
		sqltypes.Float64,
		sqltypes.Timestamp,
		sqltypes.Date,
		sqltypes.Time,
		sqltypes.Datetime,
		sqltypes.Year,
		sqltypes.Decimal,
		sqltypes.Text,
		sqltypes.Blob,
		sqltypes.VarChar,
		sqltypes.VarBinary,
		sqltypes.Char,
		sqltypes.Binary,
		sqltypes.Bit,
		sqltypes.Enum,
		sqltypes.Set,
		sqltypes.Geometry,
		sqltypes.TypeJSON,
		sqltypes.Expression,
		sqltypes.HexNum,
		sqltypes.HexVal,
		sqltypes.Tuple,
	}

	for _, f := range funcs {
		var match []string
		for _, tt := range types {
			if f.f(tt) {
				match = append(match, tt.String())
			}
		}
		t.Logf("%s(): %s", f.name, strings.Join(match, ", "))
	}
}
