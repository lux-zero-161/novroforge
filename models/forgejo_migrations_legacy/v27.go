// Copyright 2024 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations_legacy

import (
	"novroforge.xyz/modules/timeutil"

	"code.novroforge.xyz/xorm/xorm"
)

func AddCreatedUnixToRedirect(x *xorm.Engine) error {
	type UserRedirect struct {
		ID          int64              `xorm:"pk autoincr"`
		CreatedUnix timeutil.TimeStamp `xorm:"created NOT NULL DEFAULT 0"`
	}
	return x.Sync(new(UserRedirect))
}
