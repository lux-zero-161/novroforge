// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package wings

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"novroforge.xyz/models/db"
	user_model "novroforge.xyz/models/user"
	"novroforge.xyz/modules/timeutil"
)

var (
	ErrInvalidAccessUser  = errors.New("server access requires an active individual WebMain user")
	ErrServerAccessDenied = errors.New("user does not have access to this Wings server")
)

// WingsServerAccess grants one WebMain user narrowly-scoped access to a server.
// Status and recent logs are readable by assigned users; power and console
// commands, file browsing, previews, downloads, and directory creation require explicit permissions.
type WingsServerAccess struct {
	ID               int64              `xorm:"pk autoincr"`
	ServerID         int64              `xorm:"UNIQUE(server_user) INDEX NOT NULL"`
	UserID           int64              `xorm:"UNIQUE(server_user) INDEX NOT NULL"`
	CanPower         bool               `xorm:"NOT NULL DEFAULT false"`
	CanConsole       bool               `xorm:"NOT NULL DEFAULT false"`
	CanFiles         bool               `xorm:"NOT NULL DEFAULT false"`
	CanReadFiles     bool               `xorm:"NOT NULL DEFAULT false"`
	CanDownloadFiles bool               `xorm:"NOT NULL DEFAULT false"`
	CanManageFiles   bool               `xorm:"NOT NULL DEFAULT false"`
	CreatedUnix      timeutil.TimeStamp `xorm:"created"`
}

// ServerUserListItem is the non-sensitive admin display of assigned access.
type ServerUserListItem struct {
	ID               int64
	UserID           int64
	Username         string
	CanPower         bool
	CanConsole       bool
	CanFiles         bool
	CanReadFiles     bool
	CanDownloadFiles bool
	CanManageFiles   bool
}

// UserServerListItem contains only server data visible to an assigned user.
type UserServerListItem struct {
	ID               int64
	Name             string
	UUID             string
	CanPower         bool
	CanConsole       bool
	CanFiles         bool
	CanReadFiles     bool
	CanDownloadFiles bool
	CanManageFiles   bool
}

func init() {
	db.RegisterModel(new(WingsServerAccess))
}

// GrantServerAccess assigns an existing active individual account to a server.
func GrantServerAccess(ctx context.Context, serverID int64, username string, canPower, canConsole, canFiles, canReadFiles, canDownloadFiles, canManageFiles bool) error {
	if serverID <= 0 {
		return errors.New("invalid Wings server ID")
	}
	if _, exists, err := GetServerByID(ctx, serverID); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("Wings server %d does not exist", serverID)
	}
	user, err := user_model.GetUserByName(ctx, strings.TrimSpace(username))
	if err != nil {
		if user_model.IsErrUserNotExist(err) {
			return fmt.Errorf("%w: %v", ErrInvalidAccessUser, err)
		}
		return fmt.Errorf("lookup WebMain user %q: %w", username, err)
	}
	if !user.IsIndividual() || !user.IsActive || user.ProhibitLogin {
		return ErrInvalidAccessUser
	}

	engine := db.GetEngine(ctx)
	var access WingsServerAccess
	has, err := engine.Where("server_id = ? AND user_id = ?", serverID, user.ID).Get(&access)
	if err != nil {
		return fmt.Errorf("find Wings server access: %w", err)
	}
	access.ServerID = serverID
	access.UserID = user.ID
	access.CanPower = canPower
	access.CanConsole = canConsole
	access.CanFiles = canFiles
	access.CanReadFiles = canReadFiles
	access.CanDownloadFiles = canDownloadFiles
	access.CanManageFiles = canManageFiles
	if has {
		if _, err := engine.ID(access.ID).Cols("can_power", "can_console", "can_files", "can_read_files", "can_download_files", "can_manage_files").Update(&access); err != nil {
			return fmt.Errorf("update Wings server access: %w", err)
		}
		return nil
	}
	if _, err := engine.Insert(&access); err != nil {
		return fmt.Errorf("grant Wings server access: %w", err)
	}
	return nil
}

// ListServerAccess returns assigned users grouped by server without credentials.
func ListServerAccess(ctx context.Context, serverIDs []int64) (map[int64][]ServerUserListItem, error) {
	if len(serverIDs) == 0 {
		return map[int64][]ServerUserListItem{}, nil
	}
	var accessRows []*WingsServerAccess
	if err := db.GetEngine(ctx).In("server_id", serverIDs).OrderBy("id ASC").Find(&accessRows); err != nil {
		return nil, fmt.Errorf("list Wings server access: %w", err)
	}
	userIDs := make([]int64, 0, len(accessRows))
	for _, row := range accessRows {
		userIDs = append(userIDs, row.UserID)
	}
	users, err := user_model.GetUserByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("load assigned Wings users: %w", err)
	}
	usernames := make(map[int64]string, len(users))
	for _, user := range users {
		usernames[user.ID] = user.Name
	}
	items := make(map[int64][]ServerUserListItem, len(serverIDs))
	for _, row := range accessRows {
		username := usernames[row.UserID]
		if username == "" {
			username = fmt.Sprintf("user #%d", row.UserID)
		}
		items[row.ServerID] = append(items[row.ServerID], ServerUserListItem{
			ID:               row.ID,
			UserID:           row.UserID,
			Username:         username,
			CanPower:         row.CanPower,
			CanConsole:       row.CanConsole,
			CanFiles:         row.CanFiles,
			CanReadFiles:     row.CanReadFiles,
			CanDownloadFiles: row.CanDownloadFiles,
			CanManageFiles:   row.CanManageFiles,
		})
	}
	return items, nil
}

// ListUserServers returns only servers explicitly assigned to a user.
func ListUserServers(ctx context.Context, userID int64) ([]UserServerListItem, error) {
	var accessRows []*WingsServerAccess
	if err := db.GetEngine(ctx).Where("user_id = ?", userID).OrderBy("id ASC").Find(&accessRows); err != nil {
		return nil, fmt.Errorf("list user's Wings server access: %w", err)
	}
	serverIDs := make([]int64, 0, len(accessRows))
	accessByServer := make(map[int64]*WingsServerAccess, len(accessRows))
	for _, access := range accessRows {
		serverIDs = append(serverIDs, access.ServerID)
		accessByServer[access.ServerID] = access
	}
	if len(serverIDs) == 0 {
		return []UserServerListItem{}, nil
	}
	var servers []*WingsServer
	if err := db.GetEngine(ctx).In("id", serverIDs).Find(&servers); err != nil {
		return nil, fmt.Errorf("load assigned Wings servers: %w", err)
	}
	items := make([]UserServerListItem, 0, len(servers))
	for _, server := range servers {
		access := accessByServer[server.ID]
		items = append(items, UserServerListItem{
			ID:               server.ID,
			Name:             server.Name,
			UUID:             server.UUID,
			CanPower:         access.CanPower,
			CanConsole:       access.CanConsole,
			CanFiles:         access.CanFiles,
			CanReadFiles:     access.CanReadFiles,
			CanDownloadFiles: access.CanDownloadFiles,
			CanManageFiles:   access.CanManageFiles,
		})
	}
	return items, nil
}

// GetServerAccess verifies explicit assignment and its permissions.
func GetServerAccess(ctx context.Context, serverID, userID int64) (*WingsServerAccess, bool, error) {
	var access WingsServerAccess
	exists, err := db.GetEngine(ctx).Where("server_id = ? AND user_id = ?", serverID, userID).Get(&access)
	if err != nil {
		return nil, false, fmt.Errorf("get Wings server access: %w", err)
	}
	return &access, exists, nil
}

// RevokeServerAccess removes one user's access without changing the Wings server.
func RevokeServerAccess(ctx context.Context, serverID, accessID int64) error {
	if serverID <= 0 || accessID <= 0 {
		return errors.New("invalid Wings server access ID")
	}
	affected, err := db.GetEngine(ctx).ID(accessID).Where("server_id = ?", serverID).Delete(&WingsServerAccess{})
	if err != nil {
		return fmt.Errorf("revoke Wings server access %d: %w", accessID, err)
	}
	if affected == 0 {
		return fmt.Errorf("Wings server access %d does not exist", accessID)
	}
	return nil
}
