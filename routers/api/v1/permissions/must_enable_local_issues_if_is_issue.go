// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	issues_model "novroforge.xyz/models/issues"
	"novroforge.xyz/models/unit"
)

func MustEnableLocalIssuesIfIsIssue(ctx Context, issue *issues_model.Issue) {
	if ctx.Repository().UnitEnabled(ctx.Context(), unit.TypeIssues) {
		return
	}

	if !issue.IsPull {
		ctx.NotFound()
		return
	}
}
