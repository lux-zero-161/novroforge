// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package asymkey

import (
	"testing"

	"novroforge.xyz/models/unittest"

	_ "novroforge.xyz/models/actions"
	_ "novroforge.xyz/models/activities"
	_ "novroforge.xyz/models/forgefed"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
