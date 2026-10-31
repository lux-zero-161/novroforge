// Copyright 2026 The NovroForge Authors.
// SPDX-License-Identifier: GPLv3-or-later

package integration

import (
	"net/url"
	"testing"

	"novroforge.xyz/modules/setting"
	"novroforge.xyz/modules/test"
	funding_tests "novroforge.xyz/services/funding/tests"
)

func TestFundingRetrieval(t *testing.T) {
	defer test.MockVariableValue(&setting.Service.DefaultAllowCreateOrganization, true)()

	onApplicationRun(t, func(t *testing.T, url *url.URL) {
		funding_tests.FromDefaultBranch(t)
	})
}
