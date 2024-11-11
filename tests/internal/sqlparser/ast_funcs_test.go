/*
Copyright 2022 The Vitess Authors.

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

package sqlparser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/usalko/sent/internal/sqlparser"
)

func TestAddQueryHint(t *testing.T) {
	tcs := []struct {
		comments  sqlparser.Comments
		queryHint string
		expected  sqlparser.Comments
		err       string
	}{
		{
			comments:  sqlparser.Comments{},
			queryHint: "",
			expected:  nil,
		},
		{
			comments:  sqlparser.Comments{},
			queryHint: "SET_VAR(aa)",
			expected:  sqlparser.Comments{"/*+ SET_VAR(aa) */"},
		},
		{
			comments:  sqlparser.Comments{"/* toto */"},
			queryHint: "SET_VAR(aa)",
			expected:  sqlparser.Comments{"/*+ SET_VAR(aa) */", "/* toto */"},
		},
		{
			comments:  sqlparser.Comments{"/* toto */", "/*+ SET_VAR(bb) */"},
			queryHint: "SET_VAR(aa)",
			expected:  sqlparser.Comments{"/*+ SET_VAR(bb) SET_VAR(aa) */", "/* toto */"},
		},
		{
			comments:  sqlparser.Comments{"/* toto */", "/*+ SET_VAR(bb) "},
			queryHint: "SET_VAR(aa)",
			err:       "Query hint comment is malformed",
		},
		{
			comments:  sqlparser.Comments{"/* toto */", "/*+ SET_VAR(bb) */", "/*+ SET_VAR(cc) */"},
			queryHint: "SET_VAR(aa)",
			err:       "Must have only one query hint",
		},
		{
			comments:  sqlparser.Comments{"/*+ SET_VAR(bb) */"},
			queryHint: "SET_VAR(bb)",
			expected:  sqlparser.Comments{"/*+ SET_VAR(bb) */"},
		},
	}

	for i, tc := range tcs {
		comments := tc.comments.Parsed()
		t.Run(fmt.Sprintf("%d %s", i, sqlparser.String(comments)), func(t *testing.T) {
			got, err := comments.AddQueryHint(tc.queryHint)
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, got)
			}
		})
	}
}
