// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"novroforge.xyz/modules/timeutil"

	"novroforge/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add Wings server registrations",
		Upgrade:     addWingsServerTable,
	})
}

func addWingsServerTable(x *xorm.Engine) error {
	type WingsServer struct {
		ID          int64              `xorm:"pk autoincr"`
		NodeID      int64              `xorm:"UNIQUE(node_uuid) INDEX NOT NULL"`
		UUID        string             `xorm:"UNIQUE(node_uuid) CHAR(36) NOT NULL"`
		Name        string             `xorm:"VARCHAR(64) NOT NULL"`
		CreatedUnix timeutil.TimeStamp `xorm:"created"`
		UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
	}

	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(WingsServer))
	return err
}
