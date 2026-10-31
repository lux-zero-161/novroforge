// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"novroforge.xyz/models/unit"
)

func MustEnableWiki(ctx Context) {
	if !ctx.Permission().CanRead(unit.TypeWiki) {
		ctx.NotFound()
		return
	}
}
