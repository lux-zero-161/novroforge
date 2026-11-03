// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"novroforge.xyz/modules/timeutil"

	"novroforge/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add Wings nests and eggs",
		Upgrade:     addWingsNestsAndEggs,
	})
}

func addWingsNestsAndEggs(x *xorm.Engine) error {
	type WingsNest struct {
		ID          int64              `xorm:"pk autoincr"`
		Name        string             `xorm:"VARCHAR(191) NOT NULL"`
		Description string             `xorm:"TEXT"`
		CreatedUnix timeutil.TimeStamp `xorm:"created"`
		UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
	}
	type WingsEgg struct {
		ID          int64              `xorm:"pk autoincr"`
		NestID      int64              `xorm:"INDEX NOT NULL"`
		Name        string             `xorm:"VARCHAR(191) NOT NULL"`
		Description string             `xorm:"TEXT"`
		DockerImage string             `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
		Startup     string             `xorm:"TEXT"`
		Config      string             `xorm:"TEXT"`
		CreatedUnix timeutil.TimeStamp `xorm:"created"`
		UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
	}

	if _, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(WingsNest)); err != nil {
		return err
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(WingsEgg))
	return err
}