// Copyright 2023 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forgejo_v1_20

import (
	"novroforge.xyz/modules/timeutil"

	"code.forgejo.org/xorm/xorm"
)

func AddNovroForgeBlockedUser(x *xorm.Engine) error {
	type NovroForgeBlockedUser struct {
		ID          int64              `xorm:"pk autoincr"`
		BlockID     int64              `xorm:"index"`
		UserID      int64              `xorm:"index"`
		CreatedUnix timeutil.TimeStamp `xorm:"created"`
	}

	return x.Sync(new(NovroForgeBlockedUser))
}
