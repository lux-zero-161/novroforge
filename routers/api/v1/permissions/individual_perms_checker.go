// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	api "novroforge.xyz/modules/structs"
)

func IndividualPermsChecker(ctx Context) {
	// org permissions have been checked in context.OrgAssignment(), but individual permissions haven't been checked.
	if ctx.User().IsIndividual() {
		switch ctx.User().Visibility {
		case api.VisibleTypePrivate:
			if ctx.Doer() == nil || (ctx.User().ID != ctx.Doer().ID && !IsUserSiteAdmin(ctx)) {
				ctx.NotFound("Visit Project", nil)
				return
			}
		case api.VisibleTypeLimited:
			if ctx.Doer() == nil {
				ctx.NotFound("Visit Project", nil)
				return
			}
		}
	}
}
