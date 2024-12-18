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

	"github.com/usalko/prodl/internal/sql_types"
)

// TestValue builds a Value from typ and val.
// This function should only be used for testing.
func TestValue(typ sql_types.Type, val string) sql_types.Value {
	return sql_types.MakeTrusted(typ, []byte(val))
}

func split(str string) []string {
	return strings.Split(str, "|")
}
