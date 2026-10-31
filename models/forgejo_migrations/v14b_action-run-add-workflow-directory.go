// Copyright 2025 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"code.novroforge.xyz/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "Add the column workflow_directory to the tables action_run and action_schedule",
		Upgrade:     addActionRunWorkflowDirectory,
	})
}

func addActionRunWorkflowDirectory(x *xorm.Engine) error {
	type ActionRun struct {
		WorkflowDirectory string `xorm:"workflow_directory NOT NULL DEFAULT '.novroforge/workflows'"`
	}
	type ActionSchedule struct {
		WorkflowDirectory string `xorm:"workflow_directory NOT NULL DEFAULT '.novroforge/workflows'"`
	}

	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(ActionRun), new(ActionSchedule))
	return err
}
