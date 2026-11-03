// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"novroforge.xyz/modules/timeutil"

	"novroforge/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add Wings activity log",
		Upgrade:     addWingsActivity,
	})
}

func addWingsActivity(x *xorm.Engine) error {
	type WingsActivity struct {
		ID          int64              `xorm:"pk autoincr"`
		NodeID      int64              `xorm:"INDEX NOT NULL"`
		ServerID    int64              `xorm:"INDEX NOT NULL DEFAULT 0"`
		ActorID     int64              `xorm:"INDEX NOT NULL"`
		ActorName   string             `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
		Event       string             `xorm:"VARCHAR(64) NOT NULL"`
		Detail      string             `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
		CreatedUnix timeutil.TimeStamp `xorm:"created"`
	}

	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(WingsActivity))
	return err
}
