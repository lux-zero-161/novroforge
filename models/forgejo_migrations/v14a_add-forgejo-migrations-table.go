// Copyright 2025 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"novroforge.xyz/modules/timeutil"

	"code.forgejo.org/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add forgejo_migration table",
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
