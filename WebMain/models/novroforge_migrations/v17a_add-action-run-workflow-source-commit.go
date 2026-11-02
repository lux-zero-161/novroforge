// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"novroforge.xyz/modules/optional"

	"novroforge/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add WorkflowSourceCommit to ActionRun",
		Upgrade:     addWorkflowSourceCommit,
	})
}

func addWorkflowSourceCommit(x *xorm.Engine) error {
	type ActionRun struct {
		WorkflowSourceCommit optional.Option[string]
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(ActionRun))
	return err
}
