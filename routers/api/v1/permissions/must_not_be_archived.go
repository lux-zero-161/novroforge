// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"fmt"
	"net/http"
)

func MustNotBeArchived(ctx Context) {
	if ctx.Repository().IsArchived {
		ctx.Error(http.StatusLocked, "RepoArchived", fmt.Errorf("%s is archived", ctx.Repository().LogString()))
	}
}
