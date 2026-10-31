// Copyright 2024 The NovroForge Authors
// SPDX-License-Identifier: MIT

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockProtect(t *testing.T) {
	mockable := "original"
	restore := MockProtect(&mockable)
	mockable = "tainted"
	restore()
	assert.Equal(t, "original", mockable)
}
