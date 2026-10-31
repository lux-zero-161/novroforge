// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package tests

import (
	unit_model "novroforge.xyz/models/unit"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/modules/test"
)

func SaveUnits() func() {
	disabledGlobal := unit_model.DisabledRepoUnitsGet()
	restoreDisabledGlobal := func() {
		unit_model.DisabledRepoUnitsSet(disabledGlobal)
	}
	restoreDisabledRepo := test.MockProtect(&setting.Repository.DisabledRepoUnits)

	restoreDefaultGlobal := test.MockProtect(&unit_model.DefaultRepoUnits)
	restoreDefaultRepo := test.MockProtect(&setting.Repository.DefaultRepoUnits)

	restoreForkGlobal := test.MockProtect(&unit_model.DefaultForkRepoUnits)
	restoreForkRepo := test.MockProtect(&setting.Repository.DefaultForkRepoUnits)

	return func() {
		restoreDisabledGlobal()
		restoreDisabledRepo()

		restoreDefaultGlobal()
		restoreDefaultRepo()

		restoreForkGlobal()
		restoreForkRepo()
	}
}
