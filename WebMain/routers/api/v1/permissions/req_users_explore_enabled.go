// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"novroforge.xyz/modules/setting"
)

func ReqUsersExploreEnabled(ctx Context) {
	if setting.Service.Explore.DisableUsersPage {
		ctx.NotFound()
	}
}
