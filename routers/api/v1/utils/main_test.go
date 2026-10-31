// Copyright 2018 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package utils

import (
	"testing"

	"novroforge.xyz/models/unittest"
	"novroforge.xyz/modules/setting"
	webhook_service "novroforge.xyz/services/webhook"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m, &unittest.TestOptions{
		SetUp: func() error {
			setting.LoadQueueSettings()
			return webhook_service.Init()
		},
	})
}
