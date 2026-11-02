// Copyright 2024 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package novroforge_migrations_legacy

import "novroforge/xorm/xorm"

func AddHidePronounsOptionToUser(x *xorm.Engine) error {
	type User struct {
		ID                  int64 `xorm:"pk autoincr"`
		KeepPronounsPrivate bool  `xorm:"NOT NULL DEFAULT false"`
	}

	return x.Sync(&User{})
}
