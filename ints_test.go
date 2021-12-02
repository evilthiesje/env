// Copyright 2018 Philipp Brüll <pb@simia.tech>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package env_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/simia-tech/env/v2"
)

func TestInts(t *testing.T) {
	var (
		optional = env.Ints("OPTIONAL_FIELD", []int{123})
		required = env.Ints("REQUIRED_FIELD", []int{123}, env.Required())
		allowed  = env.Ints("ALLOWED_FIELD", []int{123}, env.AllowedValues("123", "456"))
	)

	testFn := func(field *env.IntsField, value string, expectValue []int, expectErr error) func(*testing.T) {
		return func(t *testing.T) {
			require.NoError(t, os.Setenv(field.Name(), value))

			value, err := field.Get()
			if expectErr == nil {
				require.NoError(t, err)
				assert.Equal(t, expectValue, value)
			} else {
				assert.ErrorIs(t, err, expectErr)
			}
		}
	}

	t.Run("Value", testFn(optional, "456", []int{456}, nil))
	t.Run("DefaultValue", testFn(optional, "", []int{123}, nil))
	t.Run("RequiredAndSet", testFn(required, "456", []int{456}, nil))
	t.Run("RequiredNotSet", testFn(required, "", []int{123}, env.ErrRequiredValueIsMissing))
	t.Run("AllowedValue", testFn(allowed, "456", []int{456}, nil))
	t.Run("UnallowedValue", testFn(allowed, "789", []int{123}, env.ErrValueIsNotAllowed))
}
