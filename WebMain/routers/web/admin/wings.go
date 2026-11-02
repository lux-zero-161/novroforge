// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package admin

import (
	"errors"
	"net/http"
	"strconv"

	wings_model "novroforge.xyz/models/wings"
	"novroforge.xyz/modules/base"
	"novroforge.xyz/modules/log"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/services/context"
	wings_service "novroforge.xyz/services/wings"
)

const tplWingsNodes base.TplName = "admin/wings_nodes"

func WingsNodes(ctx *context.Context) {
	nodes, err := wings_model.ListNodes(ctx)
	if err != nil {
		ctx.ServerError("ListWingsNodes", err)
		return
	}

	ctx.Data["Title"] = ctx.Tr("admin.wings_nodes")
	ctx.Data["PageIsAdminWingsNodes"] = true
	ctx.Data["Nodes"] = nodes
	servers, err := wings_model.ListServers(ctx)
	if err != nil {
		ctx.ServerError("ListWingsServers", err)
		return
	}
	ctx.Data["Servers"] = servers
	ctx.HTML(http.StatusOK, tplWingsNodes)
}

func CreateWingsNode(ctx *context.Context) {
	err := wings_model.CreateNode(ctx, ctx.FormString("name"), ctx.FormString("url"), ctx.FormString("token"))
	if errors.Is(err, wings_model.ErrInvalidNodeName) {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_name"))
	} else if errors.Is(err, wings_model.ErrInvalidNodeToken) {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_token"))
	} else if errors.Is(err, wings_model.ErrInvalidNodeURL) {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_url"))
	} else if err != nil {
		ctx.ServerError("CreateWingsNode", err)
		return
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_nodes.created"))
	}

	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

func DeleteWingsNode(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	if err := wings_model.DeleteNode(ctx, id); err != nil {
		if errors.Is(err, wings_model.ErrNodeHasServers) {
			ctx.Flash.Error(ctx.Tr("admin.wings_nodes.has_servers"))
			ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
			return
		}
		log.Error("DeleteWingsNode[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.delete_failed"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}

	ctx.Flash.Success(ctx.Tr("admin.wings_nodes.deleted"))
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

func CreateWingsServer(ctx *context.Context) {
	nodeID, err := strconv.ParseInt(ctx.FormString("node_id"), 10, 64)
	if err != nil || nodeID <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	err = wings_model.CreateServer(ctx, nodeID, ctx.FormString("server_name"), ctx.FormString("server_uuid"))
	switch {
	case errors.Is(err, wings_model.ErrInvalidServerName):
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.invalid_name"))
	case errors.Is(err, wings_model.ErrInvalidServerUUID):
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.invalid_uuid"))
	case err != nil:
		log.Error("CreateWingsServer[node=%d]: %v", nodeID, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.register_failed"))
	default:
		ctx.Flash.Success(ctx.Tr("admin.wings_servers.registered_success"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

func DeleteWingsServer(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	if err := wings_model.DeleteServerRegistration(ctx, id); err != nil {
		log.Error("DeleteWingsServer[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.remove_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_servers.removed"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

func TestWingsServer(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	server, exists, err := wings_model.GetServerByID(ctx, id)
	if err != nil {
		ctx.ServerError("GetWingsServer", err)
		return
	}
	if !exists {
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.not_found"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	client, err := server.GetServerClient(ctx)
	if err == nil {
		var details wings_service.ServerDetails
		details, err = client.GetServer(ctx, server.UUID)
		if err == nil {
			ctx.Flash.Info(ctx.Tr("admin.wings_servers.status", server.Name, details.State))
			ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
			return
		}
	}
	log.Error("TestWingsServer[%d]: %v", id, err)
	ctx.Flash.Error(ctx.Tr("admin.wings_servers.status_failed"))
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

func ControlWingsServer(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	action := ctx.FormString("action")
	if action != "start" && action != "stop" && action != "restart" {
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.invalid_action"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	server, exists, err := wings_model.GetServerByID(ctx, id)
	if err != nil {
		ctx.ServerError("GetWingsServer", err)
		return
	}
	if !exists {
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.not_found"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	client, err := server.GetServerClient(ctx)
	if err == nil {
		err = client.Power(ctx, server.UUID, action)
	}
	if err != nil {
		log.Error("ControlWingsServer[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.action_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_servers.action_sent"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

func SendWingsServerCommand(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	command := ctx.FormString("command")
	server, exists, err := wings_model.GetServerByID(ctx, id)
	if err != nil {
		ctx.ServerError("GetWingsServer", err)
		return
	}
	if !exists {
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.not_found"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	client, err := server.GetServerClient(ctx)
	if err == nil {
		err = client.Commands(ctx, server.UUID, []string{command})
	}
	if err != nil {
		log.Error("SendWingsServerCommand[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.command_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_servers.command_sent"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

func TestWingsNode(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	node, exists, err := wings_model.GetNodeByID(ctx, id)
	if err != nil {
		ctx.ServerError("GetWingsNode", err)
		return
	}
	if !exists {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.not_found"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	client, err := node.GetClient()
	if err == nil {
		_, err = client.GetSystem(ctx)
	}
	if err != nil {
		log.Error("TestWingsNode[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.connection_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_nodes.connection_succeeded"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}
