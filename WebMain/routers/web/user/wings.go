// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package user

import (
	"errors"
	"net/http"
	"strconv"

	wings_model "novroforge.xyz/models/wings"
	"novroforge.xyz/modules/base"
	"novroforge.xyz/modules/log"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/services/context"
)

const (
	tplWingsServers base.TplName = "wings/servers"
	tplWingsServer  base.TplName = "wings/server"
)

func WingsServers(ctx *context.Context) {
	servers, err := wings_model.ListUserServers(ctx, ctx.Doer.ID)
	if err != nil {
		ctx.ServerError("ListUserWingsServers", err)
		return
	}
	ctx.Data["Title"] = ctx.Tr("wings.servers")
	ctx.Data["PageIsWingsServers"] = true
	ctx.Data["Servers"] = servers
	ctx.HTML(http.StatusOK, tplWingsServers)
}

func WingsServer(ctx *context.Context) {
	server, access, err := getAuthorizedWingsServer(ctx)
	if errors.Is(err, wings_model.ErrServerAccessDenied) {
		ctx.NotFound("WingsServer", nil)
		return
	}
	if err != nil {
		ctx.ServerError("GetUserWingsServer", err)
		return
	}
	client, err := server.GetServerClient(ctx)
	if err != nil {
		ctx.ServerError("GetUserWingsServerClient", err)
		return
	}
	details, err := client.GetServer(ctx, server.UUID)
	if err != nil {
		ctx.ServerError("GetUserWingsServerStatus", err)
		return
	}
	logs, err := client.Logs(ctx, server.UUID, 100)
	if err != nil {
		ctx.ServerError("GetUserWingsServerLogs", err)
		return
	}

	ctx.Data["Title"] = server.Name
	ctx.Data["PageIsWingsServers"] = true
	ctx.Data["Server"] = server
	ctx.Data["Access"] = access
	ctx.Data["Details"] = details
	ctx.Data["Logs"] = logs
	ctx.HTML(http.StatusOK, tplWingsServer)
}

func ControlWingsServer(ctx *context.Context) {
	server, access, err := getAuthorizedWingsServer(ctx)
	if err != nil {
		handleWingsServerActionError(ctx, "ControlUserWingsServer", err)
		return
	}
	if !access.CanPower {
		ctx.Error(http.StatusForbidden, ctx.Locale.TrString("wings.power_denied"))
		return
	}
	action := ctx.FormString("action")
	if action != "start" && action != "stop" && action != "restart" {
		ctx.Flash.Error(ctx.Tr("wings.invalid_action"))
		ctx.Redirect(setting.AppSubURL + "/user/servers/" + strconv.FormatInt(server.ID, 10))
		return
	}
	client, err := server.GetServerClient(ctx)
	if err == nil {
		err = client.Power(ctx, server.UUID, action)
	}
	if err != nil {
		log.Error("ControlUserWingsServer[%d]: %v", server.ID, err)
		ctx.Flash.Error(ctx.Tr("wings.action_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("wings.action_sent"))
	}
	ctx.Redirect(setting.AppSubURL + "/user/servers/" + strconv.FormatInt(server.ID, 10))
}

func SendWingsServerCommand(ctx *context.Context) {
	server, access, err := getAuthorizedWingsServer(ctx)
	if err != nil {
		handleWingsServerActionError(ctx, "SendUserWingsServerCommand", err)
		return
	}
	if !access.CanConsole {
		ctx.Error(http.StatusForbidden, ctx.Locale.TrString("wings.console_denied"))
		return
	}
	client, err := server.GetServerClient(ctx)
	if err == nil {
		err = client.Commands(ctx, server.UUID, []string{ctx.FormString("command")})
	}
	if err != nil {
		log.Error("SendUserWingsServerCommand[%d]: %v", server.ID, err)
		ctx.Flash.Error(ctx.Tr("wings.command_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr("wings.command_sent"))
	}
	ctx.Redirect(setting.AppSubURL + "/user/servers/" + strconv.FormatInt(server.ID, 10))
}

func getAuthorizedWingsServer(ctx *context.Context) (*wings_model.WingsServer, *wings_model.WingsServerAccess, error) {
	serverID, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || serverID <= 0 {
		return nil, nil, wings_model.ErrServerAccessDenied
	}
	access, exists, err := wings_model.GetServerAccess(ctx, serverID, ctx.Doer.ID)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		return nil, nil, wings_model.ErrServerAccessDenied
	}
	server, exists, err := wings_model.GetServerByID(ctx, serverID)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		return nil, nil, wings_model.ErrServerAccessDenied
	}
	return server, access, nil
}

func handleWingsServerActionError(ctx *context.Context, operation string, err error) {
	if errors.Is(err, wings_model.ErrServerAccessDenied) {
		ctx.NotFound("WingsServer", nil)
		return
	}
	log.Error("%s: %v", operation, err)
	ctx.ServerError(operation, err)
}
