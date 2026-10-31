// Copyright 2026 The NovroForge Authors.
// SPDX-License-Identifier: GPLv3-or-later

package tests

import (
	"testing"

	unit_model "novroforge.xyz/models/unit"
	apiv1_permissions "novroforge.xyz/routers/api/v1/permissions"
)

var _ = registerFunctionTest(apiv1_permissions.MustEnableWiki, functionTest{
	testCases: []*testCase{
		{
			// pass if a repository with wiki unit set is present
			data: newTestData(map[string]string{}, newSharedData().
				SetDoer().
				SetRepository(),
			),
		},
		{
			// fail if a repository is present but the wiki unit is disabled
			data: newTestData(map[string]string{}, newSharedData().
				SetDoer().
				SetRepository().
				SetRepositoryDisabledUnits([]unit_model.Type{unit_model.TypeWiki}),
			),
			error: "Not Found",
		},
	},
	interpret: func(t *testing.T, permissions *apiv1_permissions.Permissions, data *testData) {
		fixtureDisableUnits(t, permissions, data.shared.RepositoryDisabledUnits())
	},
})
