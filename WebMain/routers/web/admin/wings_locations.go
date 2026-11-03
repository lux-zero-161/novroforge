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

const tplWingsLocations base.TplName = "admin/wings_locations"

func WingsLocations(ctx *context.Context) {
	locs, err := wings_model.ListLocations(ctx)
	if err != nil {
		ctx.ServerError("ListWingsLocations", err)
		return
	}
	allocs := make([]*wings_model.WingsAllocation, 0)
	for _, loc := range locs {
		_ = loc
	}
	// Load allocations for all nodes
	nodes, _ := wings_model.ListNodes(ctx)
	for _, n := range nodes {
		a, err := wings_model.ListAllocations(ctx, n.ID)
		if err != nil {
			ctx.ServerError("ListWingsAllocations", err)
			return
		}
		allocs = append(allocs, a...)
	}
	ctx.Data["Title"] = ctx.Tr("admin.wings_locations")
	ctx.Data["PageIsAdminWingsLocations"] = true
	ctx.Data["Locations"] = locs
	ctx.Data["Nodes"] = nodes
	ctx.Data["Allocations"] = allocs
	ctx.HTML(http.StatusOK, tplWingsLocations)
}

func CreateWingsLocation(ctx *context.Context) {
	if _, err := wings_model.CreateLocation(ctx, ctx.FormString("short"), ctx.FormString("long")); err != nil {
		log.Error("CreateWingsLocation: %v", err)
		ctx.Flash.Error(ctx.Tr("admin.wings_locations.create_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_locations.created"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/locations")
}

func DeleteWingsLocation(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/locations")
		return
	}
	if err := wings_model.DeleteLocation(ctx, id); err != nil {
		log.Error("DeleteWingsLocation[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_locations.delete_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_locations.deleted"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/locations")
}

func CreateWingsAllocation(ctx *context.Context) {
	nodeID, err := strconv.ParseInt(ctx.FormString("node_id"), 10, 64)
	if err != nil || nodeID <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/locations")
		return
	}
	port, _ := strconv.Atoi(ctx.FormString("port"))
	if _, err := wings_model.CreateAllocation(ctx, nodeID, ctx.FormString("ip"), port, ctx.FormString("alias")); err != nil {
		log.Error("CreateWingsAllocation: %v", err)
		ctx.Flash.Error(ctx.Tr("admin.wings_locations.allocation_create_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_locations.allocation_created"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/locations")
}

func DeleteWingsAllocation(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/locations")
		return
	}
	if err := wings_model.DeleteAllocation(ctx, id); err != nil {
		log.Error("DeleteWingsAllocation[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_locations.allocation_delete_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_locations.allocation_deleted"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/locations")
}
