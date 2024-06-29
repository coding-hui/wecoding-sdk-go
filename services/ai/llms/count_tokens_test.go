// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package llms

import (
	"gotest.tools/assert"
	"testing"
)

func TestCountTokens(t *testing.T) {
	t.Parallel()
	numTokens := CountTokens("gpt-3.5-turbo", "test for counting tokens")
	expectedNumTokens := 4
	assert.Equal(t, expectedNumTokens, numTokens)
}
