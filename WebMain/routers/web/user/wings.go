// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package user

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	wings_model "novroforge.xyz/models/wings"
	"novroforge.xyz/modules/base"
	"novroforge.xyz/modules/log"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/services/context"
	wings_service "novroforge.xyz/services/wings"
)

const (
	tplWingsServers base.TplName = "wings/servers"
	tplWingsServer  base.TplName = "wings/server"
	tplWingsFiles   base.TplName = "wings/files"
	tplWingsPreview base.TplName = "wings/file_preview"
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

func WingsServerLogs(ctx *context.Context) {
	server, _, err := getAuthorizedWingsServer(ctx)
	if errors.Is(err, wings_model.ErrServerAccessDenied) {
		ctx.NotFound("WingsServerLogs", nil)
		return
	}
	if err != nil {
		ctx.ServerError("GetUserWingsServerLogs", err)
		return
	}
	ctx.RespHeader().Set("Cache-Control", "no-store")
	client, err := server.GetServerClient(ctx)
	if err != nil {
		ctx.ServerError("GetUserWingsServerLogClient", err)
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
	ctx.JSON(http.StatusOK, struct {
		Logs    []string                    `json:"logs"`
		Details wings_service.ServerDetails `json:"details"`
	}{
		Logs:    logs,
		Details: details,
	})
}

type wingsFileListEntry struct {
	Name         string
	ModeBits     string
	Modified     string
	Mime         string
	Size         int64
	Directory    bool
	File         bool
	Symlink      bool
	Path         string
	PreviewPath  string
	DownloadPath string
}

func WingsServerFiles(ctx *context.Context) {
	server, access, err := getAuthorizedWingsServer(ctx)
	if errors.Is(err, wings_model.ErrServerAccessDenied) {
		ctx.NotFound("WingsServerFiles", nil)
		return
	}
	if err != nil {
		ctx.ServerError("GetUserWingsServerFiles", err)
		return
	}
	if !access.CanFiles && !access.CanReadFiles && !access.CanDownloadFiles && !access.CanManageFiles {
		ctx.Error(http.StatusForbidden, ctx.Locale.TrString("wings.files_denied"))
		return
	}
	directory, err := wings_service.NormalizeDirectoryPath(ctx.Req.URL.Query().Get("directory"))
	if err != nil {
		ctx.Error(http.StatusBadRequest, ctx.Locale.TrString("wings.invalid_directory"))
		return
	}
	client, err := server.GetServerClient(ctx)
	if err != nil {
		ctx.ServerError("GetUserWingsServerFilesClient", err)
		return
	}
	files, err := client.ListDirectory(ctx, server.UUID, directory)
	if err != nil {
		if errors.Is(err, wings_service.ErrInvalidDirectory) {
			ctx.Error(http.StatusBadRequest, ctx.Locale.TrString("wings.invalid_directory"))
			return
		}
		ctx.ServerError("GetUserWingsServerFiles", err)
		return
	}

	entries := make([]wingsFileListEntry, 0, len(files))
	for _, file := range files {
		entry := wingsFileListEntry{
			Name:      file.Name,
			ModeBits:  file.ModeBits,
			Modified:  file.Modified,
			Mime:      file.Mime,
			Size:      file.Size,
			Directory: file.Directory,
			File:      file.File,
			Symlink:   file.Symlink,
		}
		if file.Directory && !file.Symlink {
			entry.Path = url.QueryEscape(path.Join(directory, file.Name))
		}
		if access.CanReadFiles && file.File && !file.Symlink {
			entry.PreviewPath = url.QueryEscape(path.Join(directory, file.Name))
		}
		if access.CanDownloadFiles && file.File && !file.Symlink {
			entry.DownloadPath = url.QueryEscape(path.Join(directory, file.Name))
		}
		entries = append(entries, entry)
	}

	parentDirectory := path.Dir(directory)
	ctx.RespHeader().Set("Cache-Control", "no-store")
	ctx.Data["Title"] = ctx.Tr("wings.files")
	ctx.Data["PageIsWingsServers"] = true
	ctx.Data["Server"] = server
	ctx.Data["Directory"] = directory
	ctx.Data["ParentDirectory"] = url.QueryEscape(parentDirectory)
	ctx.Data["HasParentDirectory"] = directory != "/"
	ctx.Data["RootDirectory"] = url.QueryEscape("/")
	ctx.Data["CanReadFiles"] = access.CanReadFiles
	ctx.Data["CanDownloadFiles"] = access.CanDownloadFiles
	ctx.Data["CanManageFiles"] = access.CanManageFiles
	ctx.Data["Files"] = entries
	ctx.HTML(http.StatusOK, tplWingsFiles)
}

func CreateWingsServerDirectory(ctx *context.Context) {
	server, access, err := getAuthorizedWingsServer(ctx)
	if errors.Is(err, wings_model.ErrServerAccessDenied) {
		ctx.NotFound("CreateWingsServerDirectory", nil)
		return
	}
	if err != nil {
		ctx.ServerError("CreateUserWingsServerDirectory", err)
		return
	}
	if !access.CanManageFiles {
		ctx.Error(http.StatusForbidden, ctx.Locale.TrString("wings.file_manage_denied"))
		return
	}

	directory, err := wings_service.NormalizeDirectoryPath(ctx.FormString("directory"))
	if err != nil {
		ctx.Flash.Error(ctx.Tr("wings.invalid_directory"))
		ctx.Redirect(wingsServerFilesURL(server.ID, ctx.FormString("directory")))
		return
	}
	name := ctx.FormString("name")
	client, err := server.GetServerClient(ctx)
	if err != nil {
		ctx.ServerError("CreateUserWingsServerDirectoryClient", err)
		return
	}
	if err := client.CreateDirectory(ctx, server.UUID, name, directory); err != nil {
		if errors.Is(err, wings_service.ErrInvalidDirectoryName) {
			ctx.Flash.Error(ctx.Tr("wings.invalid_directory_name"))
			ctx.Redirect(wingsServerFilesURL(server.ID, directory))
			return
		}
		if errors.Is(err, wings_service.ErrInvalidDirectory) {
			ctx.Flash.Error(ctx.Tr("wings.invalid_directory"))
			ctx.Redirect(wingsServerFilesURL(server.ID, directory))
			return
		}
		log.Error("CreateUserWingsServerDirectory[%d]: %v", server.ID, err)
		ctx.Flash.Error(ctx.Tr("wings.file_operation_failed"))
		ctx.Redirect(wingsServerFilesURL(server.ID, directory))
		return
	}

	ctx.Flash.Success(ctx.Tr("wings.directory_created"))
	ctx.Redirect(wingsServerFilesURL(server.ID, directory))
}

func wingsServerFilesURL(serverID int64, directory string) string {
	return setting.AppSubURL + "/user/servers/" + strconv.FormatInt(serverID, 10) + "/files?directory=" + url.QueryEscape(directory)
}

func WingsServerFileDownload(ctx *context.Context) {
	server, access, err := getAuthorizedWingsServer(ctx)
	if errors.Is(err, wings_model.ErrServerAccessDenied) {
		ctx.NotFound("WingsServerFileDownload", nil)
		return
	}
	if err != nil {
		ctx.ServerError("GetUserWingsServerFileDownload", err)
		return
	}
	if !access.CanDownloadFiles {
		ctx.Error(http.StatusForbidden, ctx.Locale.TrString("wings.file_download_denied"))
		return
	}

	filePath, err := wings_service.NormalizeDirectoryPath(ctx.Req.URL.Query().Get("file"))
	if err != nil || filePath == "/" || strings.HasSuffix(ctx.Req.URL.Query().Get("file"), "/") {
		ctx.Error(http.StatusBadRequest, ctx.Locale.TrString("wings.invalid_directory"))
		return
	}
	client, err := server.GetServerClient(ctx)
	if err != nil {
		ctx.ServerError("GetUserWingsServerFileDownloadClient", err)
		return
	}
	entries, err := client.ListDirectory(ctx, server.UUID, path.Dir(filePath))
	if err != nil {
		ctx.ServerError("GetUserWingsServerFileDownloadDirectory", err)
		return
	}
	var isRegularFile bool
	for _, entry := range entries {
		if entry.Name == path.Base(filePath) && entry.File && !entry.Symlink {
			isRegularFile = true
			break
		}
	}
	if !isRegularFile {
		ctx.NotFound("WingsServerFileDownload", nil)
		return
	}

	response, err := client.FileDownload(ctx, server.UUID, filePath)
	if err != nil {
		ctx.ServerError("GetUserWingsServerFileDownload", err)
		return
	}
	defer response.Body.Close()

	filename := mime.FormatMediaType("attachment", map[string]string{"filename": path.Base(filePath)})
	if filename == "" {
		ctx.Error(http.StatusBadRequest, ctx.Locale.TrString("wings.invalid_directory"))
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
		log.Error("GetUserWingsServerFileDownload[%d]: %v", server.ID, err)
	}
}

func WingsServerFilePreview(ctx *context.Context) {
	server, access, err := getAuthorizedWingsServer(ctx)
	if errors.Is(err, wings_model.ErrServerAccessDenied) {
		ctx.NotFound("WingsServerFilePreview", nil)
		return
	}
	if err != nil {
		ctx.ServerError("GetUserWingsServerFilePreview", err)
		return
	}
	if !access.CanReadFiles {
		ctx.Error(http.StatusForbidden, ctx.Locale.TrString("wings.file_preview_denied"))
		return
	}

	filePath := ctx.Req.URL.Query().Get("file")
	if strings.HasSuffix(filePath, "/") {
		ctx.Error(http.StatusBadRequest, ctx.Locale.TrString("wings.invalid_directory"))
		return
	}
	filePath, err = wings_service.NormalizeDirectoryPath(filePath)
	if err != nil || filePath == "/" {
		ctx.Error(http.StatusBadRequest, ctx.Locale.TrString("wings.invalid_directory"))
		return
	}
	client, err := server.GetServerClient(ctx)
	if err != nil {
		ctx.ServerError("GetUserWingsServerFilePreviewClient", err)
		return
	}
	entries, err := client.ListDirectory(ctx, server.UUID, path.Dir(filePath))
	if err != nil {
		ctx.ServerError("GetUserWingsServerFilePreviewDirectory", err)
		return
	}
	var isRegularFile bool
	for _, entry := range entries {
		if entry.Name == path.Base(filePath) && entry.File && !entry.Symlink {
			isRegularFile = true
			break
		}
	}
	if !isRegularFile {
		ctx.NotFound("WingsServerFilePreview", nil)
		return
	}

	content, err := client.FileContents(ctx, server.UUID, filePath)
	switch {
	case errors.Is(err, wings_service.ErrInvalidFilePath):
		ctx.Error(http.StatusBadRequest, ctx.Locale.TrString("wings.invalid_directory"))
		return
	case errors.Is(err, wings_service.ErrFileTooLarge):
		ctx.Error(http.StatusRequestEntityTooLarge, ctx.Locale.TrString("wings.file_preview_too_large"))
		return
	case errors.Is(err, wings_service.ErrUnsupportedFileType):
		ctx.Error(http.StatusUnsupportedMediaType, ctx.Locale.TrString("wings.file_preview_unsupported"))
		return
	case err != nil:
		ctx.ServerError("GetUserWingsServerFilePreview", err)
		return
	}

	ctx.RespHeader().Set("Cache-Control", "no-store")
	ctx.RespHeader().Set("X-Content-Type-Options", "nosniff")
	ctx.Data["Title"] = ctx.Tr("wings.file_preview")
	ctx.Data["PageIsWingsServers"] = true
	ctx.Data["Server"] = server
	ctx.Data["FilePath"] = filePath
	ctx.Data["FileContent"] = content
	ctx.Data["DirectoryURL"] = setting.AppSubURL + "/user/servers/" + strconv.FormatInt(server.ID, 10) + "/files?directory=" + url.QueryEscape(path.Dir(filePath))
	ctx.HTML(http.StatusOK, tplWingsPreview)
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
