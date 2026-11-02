// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package user

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	wings_model "novroforge.xyz/models/wings"
	"novroforge.xyz/modules/base"
	"novroforge.xyz/modules/log"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/services/context"
	wings_service "novroforge.xyz/services/wings"
)

const tplWingsBackups base.TplName = "wings/backups"

type wingsBackupStatus struct {
	UUID          string `json:"uuid"`
	Status        string `json:"status"`
	Size          string `json:"size"`
	RestoreStatus string `json:"restore_status"`
}

func WingsServerBackups(ctx *context.Context) {
	server, access, err := getAuthorizedWingsServer(ctx)
	if errors.Is(err, wings_model.ErrServerAccessDenied) {
		ctx.NotFound("WingsServerBackups", nil)
		return
	}
	if err != nil {
		ctx.ServerError("GetUserWingsServerBackups", err)
		return
	}
	if !access.CanBackups && !access.CanDownloadBackups {
		ctx.Error(http.StatusForbidden, ctx.Locale.TrString("wings.backups_denied"))
		return
	}
	backups, err := wings_model.ListServerBackups(ctx, server.ID)
	if err != nil {
		ctx.ServerError("ListUserWingsServerBackups", err)
		return
	}
	callbacksConfigured := false
	if access.CanBackups {
		callbacksConfigured, err = server.HasRemoteCallbackCredentials(ctx)
		if err != nil {
			ctx.ServerError("GetUserWingsServerBackupCredentials", err)
			return
		}
	}
	ctx.RespHeader().Set("Cache-Control", "no-store")
	ctx.Data["Title"] = ctx.Tr("wings.backups")
	ctx.Data["PageIsWingsServers"] = true
	ctx.Data["Server"] = server
	ctx.Data["Backups"] = backups
	ctx.Data["CallbacksConfigured"] = callbacksConfigured
	ctx.Data["Access"] = access
	ctx.HTML(http.StatusOK, tplWingsBackups)
}

func WingsServerBackupStatuses(ctx *context.Context) {
	server, access, err := getAuthorizedWingsServer(ctx)
	if errors.Is(err, wings_model.ErrServerAccessDenied) {
		ctx.NotFound("WingsServerBackupStatuses", nil)
		return
	}
	if err != nil {
		ctx.ServerError("GetUserWingsServerBackupStatuses", err)
		return
	}
	if !access.CanBackups && !access.CanDownloadBackups {
		ctx.Error(http.StatusForbidden, ctx.Locale.TrString("wings.backups_denied"))
		return
	}
	backups, err := wings_model.ListServerBackups(ctx, server.ID)
	if err != nil {
		ctx.ServerError("ListUserWingsServerBackupStatuses", err)
		return
	}
	statuses := make([]wingsBackupStatus, 0, len(backups))
	for _, backup := range backups {
		statuses = append(statuses, wingsBackupStatus{
			UUID:          backup.UUID,
			Status:        backup.Status,
			Size:          strconv.FormatInt(backup.Size, 10),
			RestoreStatus: backup.RestoreStatus,
		})
	}
	ctx.RespHeader().Set("Cache-Control", "no-store")
	ctx.JSON(http.StatusOK, statuses)
}

func DownloadWingsServerBackup(ctx *context.Context) {
	server, access, err := getAuthorizedWingsServer(ctx)
	if errors.Is(err, wings_model.ErrServerAccessDenied) {
		ctx.NotFound("DownloadWingsServerBackup", nil)
		return
	}
	if err != nil {
		ctx.ServerError("DownloadUserWingsServerBackup", err)
		return
	}
	if !access.CanDownloadBackups {
		ctx.Error(http.StatusForbidden, ctx.Locale.TrString("wings.backup_download_denied"))
		return
	}
	backup, exists, err := wings_model.GetBackupByUUID(ctx, ctx.Params("backup"))
	if errors.Is(err, wings_model.ErrBackupNotFound) {
		ctx.NotFound("DownloadWingsServerBackup", nil)
		return
	}
	if err != nil {
		ctx.ServerError("GetUserWingsServerBackupForDownload", err)
		return
	}
	if !exists || backup.ServerID != server.ID || backup.Status != wings_model.BackupStatusCompleted {
		ctx.NotFound("DownloadWingsServerBackup", nil)
		return
	}
	client, err := server.GetServerClient(ctx)
	if err != nil {
		ctx.ServerError("DownloadUserWingsServerBackupClient", err)
		return
	}
	userUUID := uuid.NewSHA1(uuid.NameSpaceOID, []byte("novroforge-user:"+strconv.FormatInt(ctx.Doer.ID, 10))).String()
	response, err := client.DownloadBackup(ctx, server.UUID, userUUID, backup.UUID)
	if err != nil {
		log.Error("DownloadUserWingsServerBackup[%d]: %v", server.ID, err)
		ctx.ServerError("DownloadUserWingsServerBackup", err)
		return
	}
	defer response.Body.Close()

	filename := mime.FormatMediaType("attachment", map[string]string{"filename": "backup-" + backup.UUID + ".tar.gz"})
	if filename == "" {
		ctx.Error(http.StatusInternalServerError, ctx.Locale.TrString("wings.backup_download_failed"))
		return
	}
	header := ctx.Resp.Header()
	header.Set("Content-Type", "application/octet-stream")
	header.Set("Content-Disposition", filename)
	header.Set("Cache-Control", "no-store")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Accel-Buffering", "no")
	if response.ContentLength >= 0 {
		header.Set("Content-Length", strconv.FormatInt(response.ContentLength, 10))
	}
	ctx.Resp.WriteHeader(http.StatusOK)
	if _, err := io.Copy(ctx.Resp, response.Body); err != nil {
		log.Error("DownloadUserWingsServerBackup[%d] stream failed: %v", server.ID, err)
	}
}

func CreateWingsServerBackup(ctx *context.Context) {
	server, access, err := getAuthorizedWingsServer(ctx)
	if errors.Is(err, wings_model.ErrServerAccessDenied) {
		ctx.NotFound("CreateWingsServerBackup", nil)
		return
	}
	if err != nil {
		ctx.ServerError("CreateUserWingsServerBackup", err)
		return
	}
	if !access.CanBackups {
		ctx.Error(http.StatusForbidden, ctx.Locale.TrString("wings.backups_denied"))
		return
	}
	callbacksConfigured, err := server.HasRemoteCallbackCredentials(ctx)
	if err != nil {
		ctx.ServerError("GetUserWingsServerBackupCredentials", err)
		return
	}
	if !callbacksConfigured {
		ctx.Flash.Error(ctx.Tr("wings.backup_callbacks_not_configured"))
		ctx.Redirect(wingsServerBackupsURL(server.ID))
		return
	}
	client, err := server.GetServerClient(ctx)
	if err != nil {
		ctx.ServerError("CreateUserWingsServerBackupClient", err)
		return
	}
	backup, err := wings_model.CreateBackupRecord(ctx, server.ID, ctx.FormString("name"))
	if errors.Is(err, wings_model.ErrInvalidBackupName) {
		ctx.Flash.Error(ctx.Tr("wings.backup_invalid_name"))
		ctx.Redirect(wingsServerBackupsURL(server.ID))
		return
	}
	if err != nil {
		ctx.ServerError("CreateUserWingsServerBackupRecord", err)
		return
	}
	if err := client.CreateBackup(ctx, server.UUID, backup.UUID); err != nil {
		if updateErr := wings_model.MarkBackupRequestFailed(ctx, backup.ID); updateErr != nil {
			log.Error("MarkBackupRequestFailed[%d]: %v", backup.ID, updateErr)
		}
		log.Error("CreateUserWingsServerBackup[%d]: %v", server.ID, err)
		ctx.Flash.Error(ctx.Tr("wings.backup_create_failed"))
		ctx.Redirect(wingsServerBackupsURL(server.ID))
		return
	}
	ctx.Flash.Success(ctx.Tr("wings.backup_create_started"))
	ctx.Redirect(wingsServerBackupsURL(server.ID))
}

func RestoreWingsServerBackup(ctx *context.Context) {
	server, access, err := getAuthorizedWingsServer(ctx)
	if errors.Is(err, wings_model.ErrServerAccessDenied) {
		ctx.NotFound("RestoreWingsServerBackup", nil)
		return
	}
	if err != nil {
		ctx.ServerError("RestoreUserWingsServerBackup", err)
		return
	}
	if !access.CanBackups {
		ctx.Error(http.StatusForbidden, ctx.Locale.TrString("wings.backups_denied"))
		return
	}
	if !ctx.FormBool("confirm_restore") {
		ctx.Error(http.StatusBadRequest, ctx.Locale.TrString("wings.backup_restore_confirmation_required"))
		return
	}
	callbacksConfigured, err := server.HasRemoteCallbackCredentials(ctx)
	if err != nil {
		ctx.ServerError("GetUserWingsServerBackupCredentials", err)
		return
	}
	if !callbacksConfigured {
		ctx.Flash.Error(ctx.Tr("wings.backup_callbacks_not_configured"))
		ctx.Redirect(wingsServerBackupsURL(server.ID))
		return
	}
	backup, exists, err := wings_model.GetBackupByUUID(ctx, ctx.Params("backup"))
	if errors.Is(err, wings_model.ErrBackupNotFound) {
		ctx.NotFound("RestoreWingsServerBackup", nil)
		return
	}
	if err != nil {
		ctx.ServerError("GetUserWingsServerBackup", err)
		return
	}
	if !exists || backup.ServerID != server.ID {
		ctx.NotFound("RestoreWingsServerBackup", nil)
		return
	}
	client, err := server.GetServerClient(ctx)
	if err != nil {
		ctx.ServerError("RestoreUserWingsServerBackupClient", err)
		return
	}
	backup, err = wings_model.MarkBackupRestoring(ctx, backup.UUID)
	if errors.Is(err, wings_model.ErrBackupNotReady) || errors.Is(err, wings_model.ErrBackupNotFound) {
		ctx.Flash.Error(ctx.Tr("wings.backup_not_ready"))
		ctx.Redirect(wingsServerBackupsURL(server.ID))
		return
	}
	if err != nil {
		ctx.ServerError("MarkUserWingsServerBackupRestoring", err)
		return
	}
	if err := client.RestoreBackup(ctx, server.UUID, backup.UUID); err != nil {
		var statusErr *wings_service.HTTPStatusError
		if errors.As(err, &statusErr) && statusErr.StatusCode >= http.StatusBadRequest && statusErr.StatusCode < http.StatusInternalServerError {
			if updateErr := wings_model.MarkBackupRestoreRequestFailed(ctx, backup.UUID); updateErr != nil {
				log.Error("MarkBackupRestoreRequestFailed[%d]: %v", server.ID, updateErr)
			}
		}
		log.Error("RestoreUserWingsServerBackup[%d]: %v", server.ID, err)
		ctx.Flash.Error(ctx.Tr("wings.backup_restore_failed"))
		ctx.Redirect(wingsServerBackupsURL(server.ID))
		return
	}
	ctx.Flash.Success(ctx.Tr("wings.backup_restore_started"))
	ctx.Redirect(wingsServerBackupsURL(server.ID))
}

func DeleteWingsServerBackup(ctx *context.Context) {
	server, access, err := getAuthorizedWingsServer(ctx)
	if errors.Is(err, wings_model.ErrServerAccessDenied) {
		ctx.NotFound("DeleteWingsServerBackup", nil)
		return
	}
	if err != nil {
		ctx.ServerError("DeleteUserWingsServerBackup", err)
		return
	}
	if !access.CanBackups {
		ctx.Error(http.StatusForbidden, ctx.Locale.TrString("wings.backups_denied"))
		return
	}
	if !ctx.FormBool("confirm_delete") {
		ctx.Error(http.StatusBadRequest, ctx.Locale.TrString("wings.backup_delete_confirmation_required"))
		return
	}
	backup, exists, err := wings_model.GetBackupByUUID(ctx, ctx.Params("backup"))
	if errors.Is(err, wings_model.ErrBackupNotFound) {
		ctx.NotFound("DeleteWingsServerBackup", nil)
		return
	}
	if err != nil {
		ctx.ServerError("GetUserWingsServerBackup", err)
		return
	}
	if !exists || backup.ServerID != server.ID {
		ctx.NotFound("DeleteWingsServerBackup", nil)
		return
	}
	if backup.Status == wings_model.BackupStatusPending || backup.RestoreStatus == wings_model.BackupRestorePending {
		ctx.Flash.Error(ctx.Tr("wings.backup_not_ready"))
		ctx.Redirect(wingsServerBackupsURL(server.ID))
		return
	}
	client, err := server.GetServerClient(ctx)
	if err != nil {
		ctx.ServerError("DeleteUserWingsServerBackupClient", err)
		return
	}
	if err := client.DeleteBackup(ctx, server.UUID, backup.UUID); err != nil {
		log.Error("DeleteUserWingsServerBackupArchive[%d]: %v", server.ID, err)
		ctx.Flash.Error(ctx.Tr("wings.backup_delete_failed"))
		ctx.Redirect(wingsServerBackupsURL(server.ID))
		return
	}
	if err := wings_model.DeleteBackupRecord(ctx, backup.ID); err != nil {
		log.Error("DeleteUserWingsServerBackupRecord[%d]: %v", server.ID, err)
		ctx.Flash.Error(ctx.Tr("wings.backup_delete_failed"))
		ctx.Redirect(wingsServerBackupsURL(server.ID))
		return
	}
	ctx.Flash.Success(ctx.Tr("wings.backup_deleted"))
	ctx.Redirect(wingsServerBackupsURL(server.ID))
}

func wingsServerBackupsURL(serverID int64) string {
	return setting.AppSubURL + "/user/servers/" + strconv.FormatInt(serverID, 10) + "/backups"
}
