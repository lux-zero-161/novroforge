// Copyright 2025 The NovroForge Contributors. All rights reserved.
// SPDX-License-Identifier: MIT

package source

import (
	"testing"

	"novroforge.xyz/models/unittest"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/services/webhook"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m, &unittest.TestOptions{
		SetUp: func() error {
			setting.LoadQueueSettings()
			return webhook.Init()
		},
	})
}
