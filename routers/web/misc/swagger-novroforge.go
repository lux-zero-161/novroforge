// Copyright 2017 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package misc

import (
	"net/http"

	"forgejo.org/modules/base"
	"forgejo.org/services/context"
)

// tplNovroForgeSwagger swagger page template
const tplNovroForgeSwagger base.TplName = "swagger/novroforge-ui"

func SwaggerNovroForge(ctx *context.Context) {
	ctx.Data["APIVersion"] = "v1"
	ctx.HTML(http.StatusOK, tplNovroForgeSwagger)
}
