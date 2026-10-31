// Copyright 2024 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package novroforge_migrations_legacy

import "code.novroforge.xyz/xorm/xorm"

func AddHideArchiveLinksToRelease(x *xorm.Engine) error {
	type Release struct {
		ID               int64 `xorm:"pk autoincr"`
		HideArchiveLinks bool  `xorm:"NOT NULL DEFAULT false"`
	}

	return x.Sync(&Release{})
}
