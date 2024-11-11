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

	"github.com/usalko/sent/internal/sqltypes"
)

// TestValue builds a Value from typ and val.
// This function should only be used for testing.
func TestValue(typ sqltypes.Type, val string) sqltypes.Value {
	return sqltypes.MakeTrusted(typ, []byte(val))
}

func split(str string) []string {
	return strings.Split(str, "|")
}
