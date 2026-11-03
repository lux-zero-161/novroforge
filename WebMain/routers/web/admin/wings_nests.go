// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package admin

import (
	"net/http"
	"strconv"

	wings_model "novroforge.xyz/models/wings"
	"novroforge.xyz/modules/base"
	"novroforge.xyz/modules/log"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/services/context"
)

const tplWingsNests base.TplName = "admin/wings_nests"

func WingsNests(ctx *context.Context) {
	nests, err := wings_model.ListNests(ctx)
	if err != nil {
		ctx.ServerError("ListWingsNests", err)
		return
	}
	eggs := make([]*wings_model.WingsEgg, 0)
	for _, nest := range nests {
		nestEggs, err := wings_model.ListEggs(ctx, nest.ID)
		if err != nil {
			ctx.ServerError("ListWingsEggs", err)
			return
		}
		eggs = append(eggs, nestEggs...)
	}
	ctx.Data["Title"] = ctx.Tr("admin.wings_nests")
	ctx.Data["PageIsAdminWingsNests"] = true
	ctx.Data["Nests"] = nests
	ctx.Data["Eggs"] = eggs
	ctx.HTML(http.StatusOK, tplWingsNests)
}

func CreateWingsNest(ctx *context.Context) {
	name := ctx.FormString("name")
	description := ctx.FormString("description")
	if _, err := wings_model.CreateNest(ctx, name, description); err != nil {
		log.Error("CreateWingsNest: %v", err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.create_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_nests.created"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
}

func DeleteWingsNest(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
		return
	}
	if err := wings_model.DeleteNest(ctx, id); err != nil {
		log.Error("DeleteWingsNest[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.delete_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_nests.deleted"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
}

func CreateWingsEgg(ctx *context.Context) {
	nestID, err := strconv.ParseInt(ctx.FormString("nest_id"), 10, 64)
	if err != nil || nestID <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.invalid_nest"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
		return
	}
	_, err = wings_model.CreateEgg(ctx, nestID, ctx.FormString("name"), ctx.FormString("description"),
		ctx.FormString("docker_image"), ctx.FormString("startup"), wings_model.EggConfig{})
	if err != nil {
		log.Error("CreateWingsEgg: %v", err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.egg_create_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_nests.egg_created"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
}

func DeleteWingsEgg(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
		return
	}
	if err := wings_model.DeleteEgg(ctx, id); err != nil {
		log.Error("DeleteWingsEgg[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nests.egg_delete_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_nests.egg_deleted"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nests")
}
