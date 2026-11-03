// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"novroforge.xyz/modules/timeutil"

	"novroforge/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add Wings backup lock flag",
		Upgrade:     addWingsBackupLock,
	})
}

func addWingsBackupLock(x *xorm.Engine) error {
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
		Locked        bool               `xorm:"NOT NULL DEFAULT false"`
		CreatedUnix   timeutil.TimeStamp `xorm:"created"`
		UpdatedUnix   timeutil.TimeStamp `xorm:"updated"`
	}

	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(WingsBackup))
	return err
}
