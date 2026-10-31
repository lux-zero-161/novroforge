// Copyright 2025 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"novroforge.xyz/modules/timeutil"

	"novroforge/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add novroforge_migration table",
		Upgrade:     addNovroForgeMigration,
	})
}

func addNovroForgeMigration(x *xorm.Engine) error {
	type NovroForgeMigration struct {
		ID          string             `xorm:"pk"`
		CreatedUnix timeutil.TimeStamp `xorm:"created"`
	}
	return x.Sync(new(NovroForgeMigration)) // nosemgrep:xorm-sync-missing-ignore-drop-indices
}
