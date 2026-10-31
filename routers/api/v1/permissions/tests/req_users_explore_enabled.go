// Copyright 2026 The NovroForge Authors.
// SPDX-License-Identifier: GPLv3-or-later

package tests

import (
	"testing"

	"novroforge.xyz/modules/setting"
	apiv1_permissions "novroforge.xyz/routers/api/v1/permissions"
)

var _ = registerFunctionTest(apiv1_permissions.ReqUsersExploreEnabled, functionTest{
	testCases: []*testCase{
		{
			// pass because the Service.Explore.DisableUsersPage setting is false by default
			data: newTestData(map[string]string{}, newSharedData()),
		},
		{
			// fail because the Service.Explore.DisableUsersPage setting is true
			data: newTestData(map[string]string{
				"Service.Explore.DisableUsersPage": "true",
			}, newSharedData()),
			error: "Not Found",
		},
	},
	protectSettingsBool: []*bool{
		&setting.Service.Explore.DisableUsersPage,
	},
	interpret: func(t *testing.T, permissions *apiv1_permissions.Permissions, data *testData) {
		setting.Service.Explore.DisableUsersPage = data.Get("Service.Explore.DisableUsersPage") == "true"
	},
})
