// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_21

import (
	"novroforge/xorm/xorm"
)

func DropDeletedBranchTable(x *xorm.Engine) error {
	return x.DropTables("deleted_branch")
}
