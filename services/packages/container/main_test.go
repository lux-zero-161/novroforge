// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package container

import (
	"testing"

	"novroforge.xyz/models/unittest"

	_ "novroforge.xyz/models"
	_ "novroforge.xyz/models/actions"
	_ "novroforge.xyz/models/activities"
	_ "novroforge.xyz/models/forgefed"
	_ "novroforge.xyz/models/packages"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
