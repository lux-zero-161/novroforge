// Copyright 2024 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package novroforge_migrations_legacy

import "code.novroforge.xyz/xorm/xorm"

func AddNormalizedFederatedURIToUser(x *xorm.Engine) error {
	type User struct {
		ID                     int64 `xorm:"pk autoincr"`
		NormalizedFederatedURI string
	}
	return x.Sync(&User{})
}
