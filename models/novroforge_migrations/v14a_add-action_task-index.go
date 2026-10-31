// Copyright 2025 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"novroforge/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add index to action_task.job_id",
		Upgrade:     addActionTaskJobIDIndex,
	})
}

func addActionTaskJobIDIndex(x *xorm.Engine) error {
	type ActionTask struct {
		JobID int64 `xorm:"index"`
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(ActionTask))
	return err
}
