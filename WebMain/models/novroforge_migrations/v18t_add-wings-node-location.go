// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"novroforge/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add location to Wings node",
		Upgrade:     addWingsNodeLocation,
	})
}

func addWingsNodeLocation(x *xorm.Engine) error {
	type WingsNode struct {
		ID         int64  `xorm:"pk autoincr"`
		LocationID int64  `xorm:"INDEX NOT NULL DEFAULT 0"`
		Name       string `xorm:"UNIQUE NOT NULL VARCHAR(64)"`
		URL        string `xorm:"TEXT NOT NULL"`
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(WingsNode))
	return err
}