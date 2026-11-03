// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"novroforge.xyz/modules/timeutil"

	"novroforge/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add Wings locations and allocations",
		Upgrade:     addWingsLocationsAndAllocations,
	})
}

func addWingsLocationsAndAllocations(x *xorm.Engine) error {
	type WingsLocation struct {
		ID          int64              `xorm:"pk autoincr"`
		Short       string             `xorm:"VARCHAR(191) NOT NULL"`
		Long        string             `xorm:"TEXT"`
		CreatedUnix timeutil.TimeStamp `xorm:"created"`
		UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
	}
	type WingsAllocation struct {
		ID          int64              `xorm:"pk autoincr"`
		NodeID      int64              `xorm:"INDEX NOT NULL"`
		IP          string             `xorm:"VARCHAR(45) NOT NULL"`
		Port        int                `xorm:"NOT NULL"`
		IPAlias     string             `xorm:"VARCHAR(255)"`
		Assigned    bool               `xorm:"NOT NULL DEFAULT false"`
		ServerID    int64              `xorm:"INDEX"`
		CreatedUnix timeutil.TimeStamp `xorm:"created"`
		UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
	}
	if _, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(WingsLocation)); err != nil {
		return err
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(WingsAllocation))
	return err
}