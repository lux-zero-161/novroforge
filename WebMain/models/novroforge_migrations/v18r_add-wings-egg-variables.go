// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"novroforge.xyz/modules/timeutil"

	"novroforge/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add Wings egg variables",
		Upgrade:     addWingsEggVariables,
	})
}

func addWingsEggVariables(x *xorm.Engine) error {
	type WingsEggVariable struct {
		ID           int64              `xorm:"pk autoincr"`
		EggID        int64              `xorm:"INDEX NOT NULL"`
		Name         string             `xorm:"VARCHAR(191) NOT NULL"`
		Description  string             `xorm:"TEXT"`
		EnvVariable  string             `xorm:"VARCHAR(191) NOT NULL"`
		DefaultValue string             `xorm:"TEXT"`
		UserViewable bool               `xorm:"NOT NULL DEFAULT true"`
		UserEditable bool               `xorm:"NOT NULL DEFAULT true"`
		Rules        string             `xorm:"TEXT"`
		SortOrder    int                `xorm:"NOT NULL DEFAULT 0"`
		CreatedUnix  timeutil.TimeStamp `xorm:"created"`
		UpdatedUnix  timeutil.TimeStamp `xorm:"updated"`
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(WingsEggVariable))
	return err
}