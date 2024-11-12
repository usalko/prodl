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

package sql_types

import (
	"strings"
	"testing"

	"github.com/usalko/sent/internal/sql_types"
)

func TestTypeValues(t *testing.T) {
	testcases := []struct {
		defined  sql_types.Type
		expected int
	}{{
		defined:  sql_types.Null,
		expected: 0,
	}, {
		defined:  sql_types.Int8,
		expected: 1 | sql_types.FlagIsIntegral,
	}, {
		defined:  sql_types.Uint8,
		expected: 2 | sql_types.FlagIsIntegral | sql_types.FlagIsUnsigned,
	}, {
		defined:  sql_types.Int16,
		expected: 3 | sql_types.FlagIsIntegral,
	}, {
		defined:  sql_types.Uint16,
		expected: 4 | sql_types.FlagIsIntegral | sql_types.FlagIsUnsigned,
	}, {
		defined:  sql_types.Int24,
		expected: 5 | sql_types.FlagIsIntegral,
	}, {
		defined:  sql_types.Uint24,
		expected: 6 | sql_types.FlagIsIntegral | sql_types.FlagIsUnsigned,
	}, {
		defined:  sql_types.Int32,
		expected: 7 | sql_types.FlagIsIntegral,
	}, {
		defined:  sql_types.Uint32,
		expected: 8 | sql_types.FlagIsIntegral | sql_types.FlagIsUnsigned,
	}, {
		defined:  sql_types.Int64,
		expected: 9 | sql_types.FlagIsIntegral,
	}, {
		defined:  sql_types.Uint64,
		expected: 10 | sql_types.FlagIsIntegral | sql_types.FlagIsUnsigned,
	}, {
		defined:  sql_types.Float32,
		expected: 11 | sql_types.FlagIsFloat,
	}, {
		defined:  sql_types.Float64,
		expected: 12 | sql_types.FlagIsFloat,
	}, {
		defined:  sql_types.Timestamp,
		expected: 13 | sql_types.FlagIsQuoted,
	}, {
		defined:  sql_types.Date,
		expected: 14 | sql_types.FlagIsQuoted,
	}, {
		defined:  sql_types.Time,
		expected: 15 | sql_types.FlagIsQuoted,
	}, {
		defined:  sql_types.Datetime,
		expected: 16 | sql_types.FlagIsQuoted,
	}, {
		defined:  sql_types.Year,
		expected: 17 | sql_types.FlagIsIntegral | sql_types.FlagIsUnsigned,
	}, {
		defined:  sql_types.Decimal,
		expected: 18,
	}, {
		defined:  sql_types.Text,
		expected: 19 | sql_types.FlagIsQuoted | sql_types.FlagIsText,
	}, {
		defined:  sql_types.Blob,
		expected: 20 | sql_types.FlagIsQuoted | sql_types.FlagIsBinary,
	}, {
		defined:  sql_types.VarChar,
		expected: 21 | sql_types.FlagIsQuoted | sql_types.FlagIsText,
	}, {
		defined:  sql_types.VarBinary,
		expected: 22 | sql_types.FlagIsQuoted | sql_types.FlagIsBinary,
	}, {
		defined:  sql_types.Char,
		expected: 23 | sql_types.FlagIsQuoted | sql_types.FlagIsText,
	}, {
		defined:  sql_types.Binary,
		expected: 24 | sql_types.FlagIsQuoted | sql_types.FlagIsBinary,
	}, {
		defined:  sql_types.Bit,
		expected: 25 | sql_types.FlagIsQuoted,
	}, {
		defined:  sql_types.Enum,
		expected: 26 | sql_types.FlagIsQuoted,
	}, {
		defined:  sql_types.Set,
		expected: 27 | sql_types.FlagIsQuoted,
	}, {
		defined:  sql_types.Geometry,
		expected: 29 | sql_types.FlagIsQuoted,
	}, {
		defined:  sql_types.TypeJSON,
		expected: 30 | sql_types.FlagIsQuoted,
	}, {
		defined:  sql_types.Expression,
		expected: 31,
	}, {
		defined:  sql_types.HexNum,
		expected: 32 | sql_types.FlagIsText,
	}, {
		defined:  sql_types.HexVal,
		expected: 33 | sql_types.FlagIsText,
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
	alltypes := []sql_types.Type{
		sql_types.Null,
		sql_types.Int8,
		sql_types.Uint8,
		sql_types.Int16,
		sql_types.Uint16,
		sql_types.Int24,
		sql_types.Uint24,
		sql_types.Int32,
		sql_types.Uint32,
		sql_types.Int64,
		sql_types.Uint64,
		sql_types.Float32,
		sql_types.Float64,
		sql_types.Timestamp,
		sql_types.Date,
		sql_types.Time,
		sql_types.Datetime,
		sql_types.Year,
		sql_types.Decimal,
		sql_types.Text,
		sql_types.Blob,
		sql_types.VarChar,
		sql_types.VarBinary,
		sql_types.Char,
		sql_types.Binary,
		sql_types.Bit,
		sql_types.Enum,
		sql_types.Set,
		sql_types.Geometry,
		sql_types.TypeJSON,
		sql_types.Expression,
		sql_types.HexNum,
		sql_types.HexVal,
	}
	for _, typ := range alltypes {
		matched := false
		if sql_types.IsSigned(typ) {
			if !sql_types.IsIntegral(typ) {
				t.Errorf("Signed type %v is not an integral", typ)
			}
			matched = true
		}
		if sql_types.IsUnsigned(typ) {
			if !sql_types.IsIntegral(typ) {
				t.Errorf("Unsigned type %v is not an integral", typ)
			}
			if matched {
				t.Errorf("%v matched more than one category", typ)
			}
			matched = true
		}
		if sql_types.IsFloat(typ) {
			if matched {
				t.Errorf("%v matched more than one category", typ)
			}
			matched = true
		}
		if sql_types.IsQuoted(typ) {
			if matched {
				t.Errorf("%v matched more than one category", typ)
			}
			matched = true
		}
		if typ == sql_types.Null || typ == sql_types.Decimal || typ == sql_types.Expression || typ == sql_types.Bit || typ == sql_types.HexNum || typ == sql_types.HexVal {
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
	if sql_types.IsIntegral(sql_types.Null) {
		t.Error("Null: IsIntegral, must be false")
	}
	if !sql_types.IsIntegral(sql_types.Int64) {
		t.Error("Int64: !IsIntegral, must be true")
	}
	if sql_types.IsSigned(sql_types.Uint64) {
		t.Error("Uint64: IsSigned, must be false")
	}
	if !sql_types.IsSigned(sql_types.Int64) {
		t.Error("Int64: !IsSigned, must be true")
	}
	if sql_types.IsUnsigned(sql_types.Int64) {
		t.Error("Int64: IsUnsigned, must be false")
	}
	if !sql_types.IsUnsigned(sql_types.Uint64) {
		t.Error("Uint64: !IsUnsigned, must be true")
	}
	if sql_types.IsFloat(sql_types.Int64) {
		t.Error("Int64: IsFloat, must be false")
	}
	if !sql_types.IsFloat(sql_types.Float64) {
		t.Error("Uint64: !IsFloat, must be true")
	}
	if sql_types.IsQuoted(sql_types.Int64) {
		t.Error("Int64: IsQuoted, must be false")
	}
	if !sql_types.IsQuoted(sql_types.Binary) {
		t.Error("Binary: !IsQuoted, must be true")
	}
	if sql_types.IsText(sql_types.Int64) {
		t.Error("Int64: IsText, must be false")
	}
	if !sql_types.IsText(sql_types.Char) {
		t.Error("Char: !IsText, must be true")
	}
	if sql_types.IsBinary(sql_types.Int64) {
		t.Error("Int64: IsBinary, must be false")
	}
	if !sql_types.IsBinary(sql_types.Binary) {
		t.Error("Char: !IsBinary, must be true")
	}
	if !sql_types.IsNumber(sql_types.Int64) {
		t.Error("Int64: !isNumber, must be true")
	}
}

func TestTypeToMySQL(t *testing.T) {
	v, f := sql_types.TypeToMySQL(sql_types.Bit)
	if v != 16 {
		t.Errorf("Bit: %d, want 16", v)
	}
	if f != sql_types.MysqlUnsigned {
		t.Errorf("Bit flag: %x, want %x", f, sql_types.MysqlUnsigned)
	}
	v, f = sql_types.TypeToMySQL(sql_types.Date)
	if v != 10 {
		t.Errorf("Bit: %d, want 10", v)
	}
	if f != sql_types.MysqlBinary {
		t.Errorf("Bit flag: %x, want %x", f, sql_types.MysqlBinary)
	}
}

func TestMySQLToType(t *testing.T) {
	testcases := []struct {
		intype  int64
		inflags int64
		outtype sql_types.Type
	}{{
		intype:  1,
		outtype: sql_types.Int8,
	}, {
		intype:  1,
		inflags: sql_types.MysqlUnsigned,
		outtype: sql_types.Uint8,
	}, {
		intype:  2,
		outtype: sql_types.Int16,
	}, {
		intype:  2,
		inflags: sql_types.MysqlUnsigned,
		outtype: sql_types.Uint16,
	}, {
		intype:  3,
		outtype: sql_types.Int32,
	}, {
		intype:  3,
		inflags: sql_types.MysqlUnsigned,
		outtype: sql_types.Uint32,
	}, {
		intype:  4,
		outtype: sql_types.Float32,
	}, {
		intype:  5,
		outtype: sql_types.Float64,
	}, {
		intype:  6,
		outtype: sql_types.Null,
	}, {
		intype:  7,
		outtype: sql_types.Timestamp,
	}, {
		intype:  8,
		outtype: sql_types.Int64,
	}, {
		intype:  8,
		inflags: sql_types.MysqlUnsigned,
		outtype: sql_types.Uint64,
	}, {
		intype:  9,
		outtype: sql_types.Int24,
	}, {
		intype:  9,
		inflags: sql_types.MysqlUnsigned,
		outtype: sql_types.Uint24,
	}, {
		intype:  10,
		outtype: sql_types.Date,
	}, {
		intype:  11,
		outtype: sql_types.Time,
	}, {
		intype:  12,
		outtype: sql_types.Datetime,
	}, {
		intype:  13,
		outtype: sql_types.Year,
	}, {
		intype:  16,
		outtype: sql_types.Bit,
	}, {
		intype:  245,
		outtype: sql_types.TypeJSON,
	}, {
		intype:  246,
		outtype: sql_types.Decimal,
	}, {
		intype:  249,
		outtype: sql_types.Text,
	}, {
		intype:  250,
		outtype: sql_types.Text,
	}, {
		intype:  251,
		outtype: sql_types.Text,
	}, {
		intype:  252,
		outtype: sql_types.Text,
	}, {
		intype:  252,
		inflags: sql_types.MysqlBinary,
		outtype: sql_types.Blob,
	}, {
		intype:  253,
		outtype: sql_types.VarChar,
	}, {
		intype:  253,
		inflags: sql_types.MysqlBinary,
		outtype: sql_types.VarBinary,
	}, {
		intype:  254,
		outtype: sql_types.Char,
	}, {
		intype:  254,
		inflags: sql_types.MysqlBinary,
		outtype: sql_types.Binary,
	}, {
		intype:  254,
		inflags: sql_types.MysqlEnum,
		outtype: sql_types.Enum,
	}, {
		intype:  254,
		inflags: sql_types.MysqlSet,
		outtype: sql_types.Set,
	}, {
		intype:  255,
		outtype: sql_types.Geometry,
	}, {
		// Binary flag must be ignored.
		intype:  8,
		inflags: sql_types.MysqlUnsigned | sql_types.MysqlBinary,
		outtype: sql_types.Uint64,
	}, {
		// Unsigned flag must be ignored
		intype:  252,
		inflags: sql_types.MysqlUnsigned | sql_types.MysqlBinary,
		outtype: sql_types.Blob,
	}}
	for _, tcase := range testcases {
		got, err := sql_types.MySQLToType(tcase.intype, tcase.inflags)
		if err != nil {
			t.Error(err)
		}
		if got != tcase.outtype {
			t.Errorf("MySQLToType(%d, %x): %v, want %v", tcase.intype, tcase.inflags, got, tcase.outtype)
		}
	}
}

func TestTypeError(t *testing.T) {
	_, err := sql_types.MySQLToType(50, 0)
	want := "unsupported type: 50"
	if err == nil || err.Error() != want {
		t.Errorf("MySQLToType: %v, want %s", err, want)
	}
}

func TestTypeEquivalenceCheck(t *testing.T) {
	if !sql_types.AreTypesEquivalent(sql_types.Int16, sql_types.Int16) {
		t.Errorf("Int16 and Int16 are same types.")
	}
	if sql_types.AreTypesEquivalent(sql_types.Int16, sql_types.Int24) {
		t.Errorf("Int16 and Int24 are not same types.")
	}
	if !sql_types.AreTypesEquivalent(sql_types.VarChar, sql_types.VarBinary) {
		t.Errorf("VarChar in binlog and VarBinary in schema are equivalent types.")
	}
	if sql_types.AreTypesEquivalent(sql_types.VarBinary, sql_types.VarChar) {
		t.Errorf("VarBinary in binlog and VarChar in schema are not equivalent types.")
	}
	if !sql_types.AreTypesEquivalent(sql_types.Int16, sql_types.Uint16) {
		t.Errorf("Int16 in binlog and Uint16 in schema are equivalent types.")
	}
	if sql_types.AreTypesEquivalent(sql_types.Uint16, sql_types.Int16) {
		t.Errorf("Uint16 in binlog and Int16 in schema are not equivalent types.")
	}
}

func TestPrintTypeChecks(t *testing.T) {
	var funcs = []struct {
		name string
		f    func(p sql_types.Type) bool
	}{
		{"IsSigned", sql_types.IsSigned},
		{"IsFloat", sql_types.IsFloat},
		{"IsUnsigned", sql_types.IsUnsigned},
		{"IsIntegral", sql_types.IsIntegral},
		{"IsText", sql_types.IsText},
		{"IsNumber", sql_types.IsNumber},
		{"IsQuoted", sql_types.IsQuoted},
		{"IsBinary", sql_types.IsBinary},
		{"IsDate", sql_types.IsDate},
		{"IsNull", sql_types.IsNull},
	}
	var types = []sql_types.Type{
		sql_types.Null,
		sql_types.Int8,
		sql_types.Uint8,
		sql_types.Int16,
		sql_types.Uint16,
		sql_types.Int24,
		sql_types.Uint24,
		sql_types.Int32,
		sql_types.Uint32,
		sql_types.Int64,
		sql_types.Uint64,
		sql_types.Float32,
		sql_types.Float64,
		sql_types.Timestamp,
		sql_types.Date,
		sql_types.Time,
		sql_types.Datetime,
		sql_types.Year,
		sql_types.Decimal,
		sql_types.Text,
		sql_types.Blob,
		sql_types.VarChar,
		sql_types.VarBinary,
		sql_types.Char,
		sql_types.Binary,
		sql_types.Bit,
		sql_types.Enum,
		sql_types.Set,
		sql_types.Geometry,
		sql_types.TypeJSON,
		sql_types.Expression,
		sql_types.HexNum,
		sql_types.HexVal,
		sql_types.Tuple,
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
