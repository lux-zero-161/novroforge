// Copyright 2026 The Forgejo Authors.
// SPDX-License-Identifier: GPLv3-or-later

package integration

import (
	"testing"

	"novroforge.xyz/models/unittest"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/modules/test"
	apiv1 "novroforge.xyz/routers/api/v1"
	apiv1_permissions_tests "novroforge.xyz/routers/api/v1/permissions/tests"
)

func TestAPIv1Permissions(t *testing.T) {
	defer test.MockVariableValue(&setting.Service.DefaultAllowCreateOrganization, true)()
	defer test.MockVariableValue(&setting.IsInTesting, true)()
	defer test.MockVariableValue(&setting.DisableGitHooks, false)()

	unittest.PrepareTestEnv(t)

	// because setting.IsInTesting == true, it will record the
	// middleware sequence of each route it builds
	apiv1.Routes()

	apiv1_permissions_tests.APIv1Permissions(t)
}
