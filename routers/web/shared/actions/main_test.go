// Copyright 2025 The Forgejo Authors.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"testing"

	"novroforge.xyz/models/unittest"

	_ "novroforge.xyz/models"
	_ "novroforge.xyz/models/forgefed"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
