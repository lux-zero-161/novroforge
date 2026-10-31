// Copyright 2024 The NovroForge Authors c/o Codeberg e.V.. All rights reserved.
// SPDX-License-Identifier: MIT

package flags

import (
	"net/http"

	"novroforge.xyz/modules/base"
	"novroforge.xyz/modules/log"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/services/context"
)

const (
	tplRepoFlags base.TplName = "repo/flags"
)

func Manage(ctx *context.Context) {
	ctx.Data["IsRepoFlagsPage"] = true
	ctx.Data["Title"] = ctx.Tr("repo.admin.manage_flags")

	flags := map[string]bool{}
	for _, f := range setting.Repository.SettableFlags {
		flags[f] = false
	}
	repoFlags, _ := ctx.Repo.Repository.ListFlags(ctx)
	for _, f := range repoFlags {
		flags[f.Name] = true
	}

	ctx.Data["Flags"] = flags

	ctx.HTML(http.StatusOK, tplRepoFlags)
}

func ManagePost(ctx *context.Context) {
	newFlags := ctx.FormStrings("flags")

	err := ctx.Repo.Repository.ReplaceAllFlags(ctx, newFlags)
	if err != nil {
		ctx.Flash.Error(ctx.Tr("repo.admin.failed_to_replace_flags"))
		log.Error("Error replacing repository flags for repo %d: %v", ctx.Repo.Repository.ID, err)
	} else {
		ctx.Flash.Success(ctx.Tr("repo.admin.flags_replaced"))
	}

	ctx.Redirect(ctx.Repo.Repository.HTMLURL() + "/flags")
}
