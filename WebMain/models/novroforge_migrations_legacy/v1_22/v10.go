// Copyright 2024 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_22

import (
	"novroforge/xorm/xorm"
)

func AddPronounsToUser(x *xorm.Engine) error {
	type User struct {
		ID       int64 `xorm:"pk autoincr"`
		Pronouns string
	}

	return x.Sync(&User{})
}
