// Copyright 2024 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package novroforge_migrations_legacy

import "code.novroforge.xyz/xorm/xorm"

// AddDeleteBranchAfterMergeToAutoMerge: add DeleteBranchAfterMerge column, setting existing rows to false
func AddDeleteBranchAfterMergeToAutoMerge(x *xorm.Engine) error {
	type AutoMerge struct {
		ID                     int64 `xorm:"pk autoincr"`
		DeleteBranchAfterMerge bool  `xorm:"NOT NULL DEFAULT false"`
	}

	return x.Sync(&AutoMerge{})
}
