// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"net/http"

	"novroforge.xyz/modules/base"
	"novroforge.xyz/modules/util"
	"novroforge.xyz/services/context"
)

const (
	tplFindFiles base.TplName = "repo/find/files"
)

// FindFiles render the page to find repository files
func FindFiles(ctx *context.Context) {
	path := ctx.Params("*")
	ctx.Data["TreeLink"] = ctx.Repo.RepoLink + "/src/" + util.PathEscapeSegments(path)
	ctx.Data["DataLink"] = ctx.Repo.RepoLink + "/tree-list/" + util.PathEscapeSegments(path)
	ctx.HTML(http.StatusOK, tplFindFiles)
}
