// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"
)

func ReqGitHook(ctx Context) {
	if !ctx.Doer().CanEditGitHook() {
		ctx.Error(http.StatusForbidden, "", "must be allowed to edit Git hooks")
		return
	}
}
