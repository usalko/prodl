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
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/usalko/sent/internal/sqltypes"
)

const (
	InvalidNeg = "-9223372036854775809"
	MinNeg     = "-9223372036854775808"
	MinPos     = "18446744073709551615"
	InvalidPos = "18446744073709551616"
)

func TestNewValue(t *testing.T) {
	testcases := []struct {
		inType sqltypes.Type
		inVal  string
		outVal sqltypes.Value
		outErr string
	}{{
		inType: sqltypes.Null,
		inVal:  "",
		outVal: sqltypes.NULL,
	}, {
		inType: sqltypes.Int8,
		inVal:  "1",
		outVal: TestValue(sqltypes.Int8, "1"),
	}, {
		inType: sqltypes.Int16,
		inVal:  "1",
		outVal: TestValue(sqltypes.Int16, "1"),
	}, {
		inType: sqltypes.Int24,
		inVal:  "1",
		outVal: TestValue(sqltypes.Int24, "1"),
	}, {
		inType: sqltypes.Int32,
		inVal:  "1",
		outVal: TestValue(sqltypes.Int32, "1"),
	}, {
		inType: sqltypes.Int64,
		inVal:  "1",
		outVal: TestValue(sqltypes.Int64, "1"),
	}, {
		inType: sqltypes.Uint8,
		inVal:  "1",
		outVal: TestValue(sqltypes.Uint8, "1"),
	}, {
		inType: sqltypes.Uint16,
		inVal:  "1",
		outVal: TestValue(sqltypes.Uint16, "1"),
	}, {
		inType: sqltypes.Uint24,
		inVal:  "1",
		outVal: TestValue(sqltypes.Uint24, "1"),
	}, {
		inType: sqltypes.Uint32,
		inVal:  "1",
		outVal: TestValue(sqltypes.Uint32, "1"),
	}, {
		inType: sqltypes.Uint64,
		inVal:  "1",
		outVal: TestValue(sqltypes.Uint64, "1"),
	}, {
		inType: sqltypes.Float32,
		inVal:  "1.00",
		outVal: TestValue(sqltypes.Float32, "1.00"),
	}, {
		inType: sqltypes.Float64,
		inVal:  "1.00",
		outVal: TestValue(sqltypes.Float64, "1.00"),
	}, {
		inType: sqltypes.Decimal,
		inVal:  "1.00",
		outVal: TestValue(sqltypes.Decimal, "1.00"),
	}, {
		inType: sqltypes.Timestamp,
		inVal:  "2012-02-24 23:19:43",
		outVal: TestValue(sqltypes.Timestamp, "2012-02-24 23:19:43"),
	}, {
		inType: sqltypes.Date,
		inVal:  "2012-02-24",
		outVal: TestValue(sqltypes.Date, "2012-02-24"),
	}, {
		inType: sqltypes.Time,
		inVal:  "23:19:43",
		outVal: TestValue(sqltypes.Time, "23:19:43"),
	}, {
		inType: sqltypes.Datetime,
		inVal:  "2012-02-24 23:19:43",
		outVal: TestValue(sqltypes.Datetime, "2012-02-24 23:19:43"),
	}, {
		inType: sqltypes.Year,
		inVal:  "1",
		outVal: TestValue(sqltypes.Year, "1"),
	}, {
		inType: sqltypes.Text,
		inVal:  "a",
		outVal: TestValue(sqltypes.Text, "a"),
	}, {
		inType: sqltypes.Blob,
		inVal:  "a",
		outVal: TestValue(sqltypes.Blob, "a"),
	}, {
		inType: sqltypes.VarChar,
		inVal:  "a",
		outVal: TestValue(sqltypes.VarChar, "a"),
	}, {
		inType: sqltypes.Binary,
		inVal:  "a",
		outVal: TestValue(sqltypes.Binary, "a"),
	}, {
		inType: sqltypes.Char,
		inVal:  "a",
		outVal: TestValue(sqltypes.Char, "a"),
	}, {
		inType: sqltypes.Bit,
		inVal:  "1",
		outVal: TestValue(sqltypes.Bit, "1"),
	}, {
		inType: sqltypes.Enum,
		inVal:  "a",
		outVal: TestValue(sqltypes.Enum, "a"),
	}, {
		inType: sqltypes.Set,
		inVal:  "a",
		outVal: TestValue(sqltypes.Set, "a"),
	}, {
		inType: sqltypes.VarBinary,
		inVal:  "a",
		outVal: TestValue(sqltypes.VarBinary, "a"),
	}, {
		inType: sqltypes.Int64,
		inVal:  InvalidNeg,
		outErr: "out of range",
	}, {
		inType: sqltypes.Int64,
		inVal:  InvalidPos,
		outErr: "out of range",
	}, {
		inType: sqltypes.Uint64,
		inVal:  "-1",
		outErr: "invalid syntax",
	}, {
		inType: sqltypes.Uint64,
		inVal:  InvalidPos,
		outErr: "out of range",
	}, {
		inType: sqltypes.Float64,
		inVal:  "a",
		outErr: "invalid syntax",
	}, {
		inType: sqltypes.Expression,
		inVal:  "a",
		outErr: "invalid type specified for MakeValue: EXPRESSION",
	}}
	for _, tcase := range testcases {
		v, err := sqltypes.NewValue(tcase.inType, []byte(tcase.inVal))
		if tcase.outErr != "" {
			if err == nil || !strings.Contains(err.Error(), tcase.outErr) {
				t.Errorf("ValueFromBytes(%v, %v) error: %v, must contain %v", tcase.inType, tcase.inVal, err, tcase.outErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("ValueFromBytes(%v, %v) error: %v", tcase.inType, tcase.inVal, err)
			continue
		}
		if !reflect.DeepEqual(v, tcase.outVal) {
			t.Errorf("ValueFromBytes(%v, %v) = %v, want %v", tcase.inType, tcase.inVal, v, tcase.outVal)
		}
	}
}

// TestNew tests 'New' functions that are not tested
// through other code paths.
func TestNew(t *testing.T) {
	got := sqltypes.NewInt32(1)
	want := sqltypes.MakeTrusted(sqltypes.Int32, []byte("1"))
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewInt32(aa): %v, want %v", got, want)
	}

	got = sqltypes.NewVarBinary("aa")
	want = sqltypes.MakeTrusted(sqltypes.VarBinary, []byte("aa"))
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewVarBinary(aa): %v, want %v", got, want)
	}
}

func TestMakeTrusted(t *testing.T) {
	v := sqltypes.MakeTrusted(sqltypes.Null, []byte("abcd"))
	if !reflect.DeepEqual(v, sqltypes.NULL) {
		t.Errorf("MakeTrusted(Null...) = %v, want null", v)
	}
	v = sqltypes.MakeTrusted(sqltypes.Int64, []byte("1"))
	want := TestValue(sqltypes.Int64, "1")
	if !reflect.DeepEqual(v, want) {
		t.Errorf("MakeTrusted(Int64, \"1\") = %v, want %v", v, want)
	}
}

func TestIntegralValue(t *testing.T) {
	testcases := []struct {
		in     string
		outVal sqltypes.Value
		outErr string
	}{{
		in:     MinNeg,
		outVal: TestValue(sqltypes.Int64, MinNeg),
	}, {
		in:     "1",
		outVal: TestValue(sqltypes.Int64, "1"),
	}, {
		in:     MinPos,
		outVal: TestValue(sqltypes.Uint64, MinPos),
	}, {
		in:     InvalidPos,
		outErr: "out of range",
	}}
	for _, tcase := range testcases {
		v, err := sqltypes.NewIntegral(tcase.in)
		if tcase.outErr != "" {
			if err == nil || !strings.Contains(err.Error(), tcase.outErr) {
				t.Errorf("BuildIntegral(%v) error: %v, must contain %v", tcase.in, err, tcase.outErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("BuildIntegral(%v) error: %v", tcase.in, err)
			continue
		}
		if !reflect.DeepEqual(v, tcase.outVal) {
			t.Errorf("BuildIntegral(%v) = %v, want %v", tcase.in, v, tcase.outVal)
		}
	}
}

func TestInterfaceValue(t *testing.T) {
	testcases := []struct {
		in  any
		out sqltypes.Value
	}{{
		in:  nil,
		out: sqltypes.NULL,
	}, {
		in:  []byte("a"),
		out: TestValue(sqltypes.VarBinary, "a"),
	}, {
		in:  int64(1),
		out: TestValue(sqltypes.Int64, "1"),
	}, {
		in:  uint64(1),
		out: TestValue(sqltypes.Uint64, "1"),
	}, {
		in:  float64(1.2),
		out: TestValue(sqltypes.Float64, "1.2"),
	}, {
		in:  "a",
		out: TestValue(sqltypes.VarChar, "a"),
	}}
	for _, tcase := range testcases {
		v, err := sqltypes.InterfaceToValue(tcase.in)
		if err != nil {
			t.Errorf("BuildValue(%#v) error: %v", tcase.in, err)
			continue
		}
		if !reflect.DeepEqual(v, tcase.out) {
			t.Errorf("BuildValue(%#v) = %v, want %v", tcase.in, v, tcase.out)
		}
	}

	_, err := sqltypes.InterfaceToValue(make(chan bool))
	want := "unexpected"
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("BuildValue(chan): %v, want %v", err, want)
	}
}

func TestAccessors(t *testing.T) {
	v := TestValue(sqltypes.Int64, "1")
	if v.Type() != sqltypes.Int64 {
		t.Errorf("v.Type=%v, want Int64", v.Type())
	}
	if !bytes.Equal(v.Raw(), []byte("1")) {
		t.Errorf("v.Raw=%s, want 1", v.Raw())
	}
	if v.Len() != 1 {
		t.Errorf("v.Len=%d, want 1", v.Len())
	}
	if v.ToString() != "1" {
		t.Errorf("v.String=%s, want 1", v.ToString())
	}
	if v.IsNull() {
		t.Error("v.IsNull: true, want false")
	}
	if !v.IsIntegral() {
		t.Error("v.IsIntegral: false, want true")
	}
	if !v.IsSigned() {
		t.Error("v.IsSigned: false, want true")
	}
	if v.IsUnsigned() {
		t.Error("v.IsUnsigned: true, want false")
	}
	if v.IsFloat() {
		t.Error("v.IsFloat: true, want false")
	}
	if v.IsQuoted() {
		t.Error("v.IsQuoted: true, want false")
	}
	if v.IsText() {
		t.Error("v.IsText: true, want false")
	}
	if v.IsBinary() {
		t.Error("v.IsBinary: true, want false")
	}
	{
		i, err := v.ToInt64()
		if err != nil {
			t.Errorf("v.ToInt64: got error: %+v, want no error", err)
		}
		if i != 1 {
			t.Errorf("v.ToInt64=%+v, want 1", i)
		}
	}
	{
		i, err := v.ToUint64()
		if err != nil {
			t.Errorf("v.ToUint64: got error: %+v, want no error", err)
		}
		if i != 1 {
			t.Errorf("v.ToUint64=%+v, want 1", i)
		}
	}
	{
		b, err := v.ToBool()
		if err != nil {
			t.Errorf("v.ToBool: got error: %+v, want no error", err)
		}
		if !b {
			t.Errorf("v.ToBool=%+v, want true", b)
		}
	}
}

func TestAccessorsNegative(t *testing.T) {
	v := TestValue(sqltypes.Int64, "-1")
	if v.ToString() != "-1" {
		t.Errorf("v.String=%s, want -1", v.ToString())
	}
	if v.IsNull() {
		t.Error("v.IsNull: true, want false")
	}
	if !v.IsIntegral() {
		t.Error("v.IsIntegral: false, want true")
	}
	{
		i, err := v.ToInt64()
		if err != nil {
			t.Errorf("v.ToInt64: got error: %+v, want no error", err)
		}
		if i != -1 {
			t.Errorf("v.ToInt64=%+v, want -1", i)
		}
	}
	{
		if _, err := v.ToUint64(); err == nil {
			t.Error("v.ToUint64: got no error, want error")
		}
	}
	{
		if _, err := v.ToBool(); err == nil {
			t.Error("v.ToUint64: got no error, want error")
		}
	}
}

func TestToBytesAndString(t *testing.T) {
	for _, v := range []sqltypes.Value{
		sqltypes.NULL,
		TestValue(sqltypes.Int64, "1"),
		TestValue(sqltypes.Int64, "12"),
	} {
		vBytes, err := v.ToBytes()
		require.NoError(t, err)
		if b := vBytes; !bytes.Equal(b, v.Raw()) {
			t.Errorf("%v.ToBytes: %s, want %s", v, b, v.Raw())
		}
		if s := v.ToString(); s != string(v.Raw()) {
			t.Errorf("%v.ToString: %s, want %s", v, s, v.Raw())
		}
	}

	tv := TestValue(sqltypes.Expression, "aa")
	tvBytes, err := tv.ToBytes()
	require.EqualError(t, err, "expression cannot be converted to bytes")
	if b := tvBytes; b != nil {
		t.Errorf("%v.ToBytes: %s, want nil", tv, b)
	}
	if s := tv.ToString(); s != "" {
		t.Errorf("%v.ToString: %s, want \"\"", tv, s)
	}
}

func TestEncode(t *testing.T) {
	testcases := []struct {
		in       sqltypes.Value
		outSQL   string
		outASCII string
	}{{
		in:       sqltypes.NULL,
		outSQL:   "null",
		outASCII: "null",
	}, {
		in:       TestValue(sqltypes.Int64, "1"),
		outSQL:   "1",
		outASCII: "1",
	}, {
		in:       TestValue(sqltypes.VarChar, "foo"),
		outSQL:   "'foo'",
		outASCII: "'Zm9v'",
	}, {
		in:       TestValue(sqltypes.VarChar, "\x00'\"\b\n\r\t\x1A\\"),
		outSQL:   "'\\0\\'\\\"\\b\\n\\r\\t\\Z\\\\'",
		outASCII: "'ACciCAoNCRpc'",
	}, {
		in:       TestValue(sqltypes.Bit, "a"),
		outSQL:   "b'01100001'",
		outASCII: "'YQ=='",
	}}
	for _, tcase := range testcases {
		buf := &bytes.Buffer{}
		tcase.in.EncodeSQL(buf)
		if tcase.outSQL != buf.String() {
			t.Errorf("%v.EncodeSQL = %q, want %q", tcase.in, buf.String(), tcase.outSQL)
		}
		buf = &bytes.Buffer{}
		tcase.in.EncodeASCII(buf)
		if tcase.outASCII != buf.String() {
			t.Errorf("%v.EncodeASCII = %q, want %q", tcase.in, buf.String(), tcase.outASCII)
		}
	}
}

// TestEncodeMap ensures DontEscape is not escaped
func TestEncodeMap(t *testing.T) {
	if sqltypes.SQLEncodeMap[sqltypes.DontEscape] != sqltypes.DontEscape {
		t.Errorf("SQLEncodeMap[DontEscape] = %v, want %v", sqltypes.SQLEncodeMap[sqltypes.DontEscape], sqltypes.DontEscape)
	}
	if sqltypes.SQLDecodeMap[sqltypes.DontEscape] != sqltypes.DontEscape {
		t.Errorf("SQLDecodeMap[DontEscape] = %v, want %v", sqltypes.SQLEncodeMap[sqltypes.DontEscape], sqltypes.DontEscape)
	}
}
