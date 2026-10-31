// Copyright 2026 The NovroForge Authors.
// SPDX-License-Identifier: GPLv3-or-later

package tests

import (
	apiv1_permissions "novroforge.xyz/routers/api/v1/permissions"
)

var _ = registerFunctionTest(apiv1_permissions.MustNotBeArchived, functionTest{
	testCases: []*testCase{
		{
			// pass if a repository is present
			data: newTestData(map[string]string{}, newSharedData().
				SetRepository(),
			),
		},
		{
			// fail if a repository is present but is archived
			data: newTestData(map[string]string{}, newSharedData().
				SetRepository().
				SetRepositoryArchived(true),
			),
			error: "is archived",
		},
	},
})
