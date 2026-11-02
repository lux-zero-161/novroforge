// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"novroforge.xyz/modules/timeutil"

	"novroforge/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add Wings server file edit access",
		Upgrade:     addWingsServerFileEditAccess,
	})
}

func addWingsServerFileEditAccess(x *xorm.Engine) error {
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
		CreatedUnix      timeutil.TimeStamp `xorm:"created"`
	}

	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(WingsServerAccess))
	return err
}
