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
	if !prepareWingsNodesPage(ctx) {
		return
	}
	ctx.HTML(http.StatusOK, tplWingsNodes)
}

func prepareWingsNodesPage(ctx *context.Context) bool {
	nodes, err := wings_model.ListNodes(ctx)
	if err != nil {
		ctx.ServerError("ListWingsNodes", err)
		return false
	}

	ctx.Data["Title"] = ctx.Tr("admin.wings_nodes")
	ctx.Data["PageIsAdminWingsNodes"] = true
	ctx.Data["Nodes"] = nodes
	ctx.Data["Locations"] = func() []*wings_model.WingsLocation {
		locs, _ := wings_model.ListLocations(ctx)
		return locs
	}()
	ctx.Data["LogsServerID"] = int64(0)
	servers, err := wings_model.ListServers(ctx)
	if err != nil {
		ctx.ServerError("ListWingsServers", err)
		return false
	}
	ctx.Data["Servers"] = servers
	activities, err := wings_model.ListActivity(ctx, 50)
	if err != nil {
		ctx.ServerError("ListWingsActivity", err)
		return false
	}
	ctx.Data["Activities"] = activities
	return true
}

func CreateWingsNode(ctx *context.Context) {
	locationID, _ := strconv.ParseInt(ctx.FormString("location_id"), 10, 64)
	err := wings_model.CreateNode(ctx, locationID, ctx.FormString("name"), ctx.FormString("url"), ctx.FormString("token"))
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

func RotateWingsNodeToken(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	err = wings_model.RotateNodeToken(ctx, id, ctx.FormString("token"))
	switch {
	case errors.Is(err, wings_model.ErrInvalidNodeToken):
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_token"))
	case err != nil:
		log.Error("RotateWingsNodeToken[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.token_rotation_failed"))
	default:
		recordWingsActivity(ctx, id, 0, wings_model.ActivityNodeTokenRotate, "")
		ctx.Flash.Success(ctx.Tr("admin.wings_nodes.token_rotated"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

func SetWingsNodeRemoteCredentials(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	err = wings_model.SetNodeRemoteTokenID(ctx, id, ctx.FormString("token_id"))
	switch {
	case errors.Is(err, wings_model.ErrInvalidRemoteCredentials):
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_remote_credentials"))
	case err != nil:
		log.Error("SetWingsNodeRemoteCredentials[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.remote_credentials_failed"))
	default:
		recordWingsActivity(ctx, id, 0, wings_model.ActivityNodeRemoteCreds, "")
		ctx.Flash.Success(ctx.Tr("admin.wings_nodes.remote_credentials_saved"))
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
	recordWingsActivity(ctx, id, 0, wings_model.ActivityNodeDelete, "")
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
		recordWingsActivity(ctx, nodeID, 0, wings_model.ActivityRegister, ctx.FormString("server_uuid"))
		ctx.Flash.Success(ctx.Tr("admin.wings_servers.registered_success"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

func RenameWingsServer(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	err = wings_model.RenameServer(ctx, id, ctx.FormString("name"))
	switch {
	case errors.Is(err, wings_model.ErrInvalidServerName):
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.invalid_name"))
	case err != nil:
		log.Error("RenameWingsServer[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.rename_failed"))
	default:
		recordWingsServerActivity(ctx, id, wings_model.ActivityRename, ctx.FormString("name"))
		ctx.Flash.Success(ctx.Tr("admin.wings_servers.renamed"))
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
	if err := wings_model.DeleteServerRegistration(ctx, id); err != nil {
		log.Error("DeleteWingsServer[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.remove_failed"))
	} else {
		recordWingsActivity(ctx, server.NodeID, server.ID, wings_model.ActivityRemove, server.Name)
		ctx.Flash.Success(ctx.Tr("admin.wings_servers.removed"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

func GrantWingsServerAccess(ctx *context.Context) {
	serverID := ctx.ParamsInt64("id")
	if err := wings_model.GrantServerAccess(
		ctx,
		serverID,
		ctx.FormString("username"),
		ctx.FormBool("can_power"),
		ctx.FormBool("can_console"),
		ctx.FormBool("can_files"),
		ctx.FormBool("can_read_files"),
		ctx.FormBool("can_download_files"),
		ctx.FormBool("can_manage_files"),
		ctx.FormBool("can_edit_files"),
		ctx.FormBool("can_delete_files"),
		ctx.FormBool("can_upload_files"),
		ctx.FormBool("can_backups"),
		ctx.FormBool("can_download_backups"),
	); err != nil {
		if errors.Is(err, wings_model.ErrInvalidAccessUser) {
			ctx.Flash.Error(ctx.Tr("admin.wings_servers.invalid_access_user"))
		} else {
			log.Error("GrantWingsServerAccess[%d]: %v", serverID, err)
			ctx.Flash.Error(ctx.Tr("admin.wings_servers.access_failed"))
		}
	} else {
		recordWingsServerActivity(ctx, serverID, wings_model.ActivityGrantAccess, ctx.FormString("username"))
		ctx.Flash.Success(ctx.Tr("admin.wings_servers.access_granted"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

func RevokeWingsServerAccess(ctx *context.Context) {
	serverID := ctx.ParamsInt64("id")
	accessID := ctx.ParamsInt64("access_id")
	if err := wings_model.RevokeServerAccess(ctx, serverID, accessID); err != nil {
		log.Error("RevokeWingsServerAccess[server=%d,access=%d]: %v", serverID, accessID, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.access_revoke_failed"))
	} else {
		recordWingsServerActivity(ctx, serverID, wings_model.ActivityRevokeAccess, "")
		ctx.Flash.Success(ctx.Tr("admin.wings_servers.access_revoked"))
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

func WingsServerLogs(ctx *context.Context) {
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
	if err != nil {
		log.Error("GetWingsServerLogs[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.logs_failed"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	logs, err := client.Logs(ctx, server.UUID, 100)
	if err != nil {
		log.Error("GetWingsServerLogs[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.logs_failed"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	if !prepareWingsNodesPage(ctx) {
		return
	}
	ctx.Data["Logs"] = logs
	ctx.Data["LogsServerID"] = id
	ctx.HTML(http.StatusOK, tplWingsNodes)
}

func ControlWingsServer(ctx *context.Context) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	action := ctx.FormString("action")
	if action != "start" && action != "stop" && action != "restart" && action != "kill" {
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.invalid_action"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	if action == "kill" && !ctx.FormBool("confirm_kill") {
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.kill_confirmation_required"))
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
		recordWingsActivity(ctx, server.NodeID, server.ID, wings_model.ActivityPower, action)
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
		recordWingsActivity(ctx, server.NodeID, server.ID, wings_model.ActivityCommand, "")
		ctx.Flash.Success(ctx.Tr("admin.wings_servers.command_sent"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

// getWingsServerClient loads a registered server and the Wings client for its
// node. It writes the flash message and redirect itself when the lookup fails.
func getWingsServerClient(ctx *context.Context) (*wings_model.WingsServer, *wings_service.Client, bool) {
	id := ctx.ParamsInt64("id")
	if id <= 0 {
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.invalid_id"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return nil, nil, false
	}
	server, exists, err := wings_model.GetServerByID(ctx, id)
	if err != nil {
		ctx.ServerError("GetWingsServer", err)
		return nil, nil, false
	}
	if !exists {
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.not_found"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return nil, nil, false
	}
	client, err := server.GetServerClient(ctx)
	if err != nil {
		log.Error("GetWingsServerClient[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.action_failed"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return nil, nil, false
	}
	return server, client, true
}

func InstallWingsServer(ctx *context.Context) {
	server, client, ok := getWingsServerClient(ctx)
	if !ok {
		return
	}
	if err := client.InstallServer(ctx, server.UUID); err != nil {
		log.Error("InstallWingsServer[%d]: %v", server.ID, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.install_failed"))
	} else {
		recordWingsActivity(ctx, server.NodeID, server.ID, wings_model.ActivityInstall, server.Name)
		ctx.Flash.Success(ctx.Tr("admin.wings_servers.install_started"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

func ReinstallWingsServer(ctx *context.Context) {
	if !ctx.FormBool("confirm_reinstall") {
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.reinstall_confirmation_required"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	server, client, ok := getWingsServerClient(ctx)
	if !ok {
		return
	}
	if err := client.ReinstallServer(ctx, server.UUID); err != nil {
		log.Error("ReinstallWingsServer[%d]: %v", server.ID, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.reinstall_failed"))
	} else {
		recordWingsActivity(ctx, server.NodeID, server.ID, wings_model.ActivityReinstall, server.Name)
		ctx.Flash.Success(ctx.Tr("admin.wings_servers.reinstall_started"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

func SyncWingsServer(ctx *context.Context) {
	server, client, ok := getWingsServerClient(ctx)
	if !ok {
		return
	}
	if err := client.SyncServer(ctx, server.UUID); err != nil {
		log.Error("SyncWingsServer[%d]: %v", server.ID, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.sync_failed"))
	} else {
		recordWingsActivity(ctx, server.NodeID, server.ID, wings_model.ActivitySync, server.Name)
		ctx.Flash.Success(ctx.Tr("admin.wings_servers.sync_succeeded"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

func DeleteWingsServerFromNode(ctx *context.Context) {
	if !ctx.FormBool("confirm_delete") {
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.delete_on_node_confirmation_required"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	server, client, ok := getWingsServerClient(ctx)
	if !ok {
		return
	}
	if err := client.DeleteServer(ctx, server.UUID); err != nil {
		log.Error("DeleteWingsServerFromNode[%d]: %v", server.ID, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.delete_on_node_failed"))
		ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
		return
	}
	recordWingsActivity(ctx, server.NodeID, server.ID, wings_model.ActivityDelete, server.Name)
	if err := wings_model.DeleteServerRegistration(ctx, server.ID); err != nil {
		log.Error("DeleteWingsServerFromNode[%d]: remove registration: %v", server.ID, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_servers.delete_on_node_partial"))
	} else {
		ctx.Flash.Success(ctx.Tr("admin.wings_servers.delete_on_node_succeeded"))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}

// recordWingsActivity stores an audit entry without interrupting the user
// action when logging fails.
func recordWingsActivity(ctx *context.Context, nodeID, serverID int64, event, detail string) {
	if err := wings_model.RecordActivity(ctx, nodeID, serverID, ctx.Doer.ID, ctx.Doer.Name, event, detail); err != nil {
		log.Error("RecordWingsActivity(event=%s, node=%d, server=%d): %v", event, nodeID, serverID, err)
	}
}

// recordWingsServerActivity looks up the node of a registered server and logs
// an audit entry for it.
func recordWingsServerActivity(ctx *context.Context, serverID int64, event, detail string) {
	server, exists, err := wings_model.GetServerByID(ctx, serverID)
	if err != nil {
		log.Error("RecordWingsServerActivity[%d]: %v", serverID, err)
		return
	}
	if !exists {
		log.Error("RecordWingsServerActivity[%d]: server does not exist", serverID)
		return
	}
	recordWingsActivity(ctx, server.NodeID, server.ID, event, detail)
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
	var info wings_service.SystemInformation
	if err == nil {
		info, err = client.GetSystem(ctx)
	}
	if err != nil {
		log.Error("TestWingsNode[%d]: %v", id, err)
		ctx.Flash.Error(ctx.Tr("admin.wings_nodes.connection_failed"))
	} else {
		ctx.Flash.Success(ctx.Tr(
			"admin.wings_nodes.connection_succeeded",
			info.OS,
			info.Architecture,
			info.CPUCount,
			info.Version,
		))
	}
	ctx.Redirect(setting.AppSubURL + "/admin/wings/nodes")
}
