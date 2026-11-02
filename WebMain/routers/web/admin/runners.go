// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package admin

import (
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/services/context"
)

func RedirectToDefaultSetting(ctx *context.Context) {
	ctx.Redirect(setting.AppSubURL + "/admin/actions/runners")
}
