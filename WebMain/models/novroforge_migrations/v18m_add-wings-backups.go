// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"novroforge.xyz/modules/timeutil"

	"novroforge/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add Wings backup metadata and access permission",
		Upgrade:     addWingsBackups,
	})
}

func addWingsBackups(x *xorm.Engine) error {
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
		CanEditFiles     bool               `xorm:"NOT NULL DEFAULT false"`
		CanDeleteFiles   bool               `xorm:"NOT NULL DEFAULT false"`
		CanUploadFiles   bool               `xorm:"NOT NULL DEFAULT false"`
		CanBackups       bool               `xorm:"NOT NULL DEFAULT false"`
		CreatedUnix      timeutil.TimeStamp `xorm:"created"`
	}
	type WingsBackup struct {
		ID            int64              `xorm:"pk autoincr"`
		UUID          string             `xorm:"UNIQUE CHAR(36) NOT NULL"`
		ServerID      int64              `xorm:"INDEX NOT NULL"`
		Name          string             `xorm:"VARCHAR(64) NOT NULL"`
		Status        string             `xorm:"VARCHAR(16) NOT NULL DEFAULT 'pending'"`
		Checksum      string             `xorm:"VARCHAR(128)"`
		ChecksumType  string             `xorm:"VARCHAR(16)"`
		Size          int64              `xorm:"NOT NULL DEFAULT 0"`
		RestoreStatus string             `xorm:"VARCHAR(16) NOT NULL DEFAULT 'none'"`
		CreatedUnix   timeutil.TimeStamp `xorm:"created"`
		UpdatedUnix   timeutil.TimeStamp `xorm:"updated"`
	}

	if _, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(WingsServerAccess)); err != nil {
		return err
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(WingsBackup))
	return err
}
