// Copyright 2023 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package novroforge_v1_20

import (
	"novroforge.xyz/modules/timeutil"

	"code.novroforge.xyz/xorm/xorm"
)

type AuthorizationToken struct {
	ID              int64  `xorm:"pk autoincr"`
	UID             int64  `xorm:"INDEX"`
	LookupKey       string `xorm:"INDEX UNIQUE"`
	HashedValidator string
	Expiry          timeutil.TimeStamp
}

func (AuthorizationToken) TableName() string {
	return "novroforge_auth_token"
}

func CreateAuthorizationTokenTable(x *xorm.Engine) error {
	return x.Sync(new(AuthorizationToken))
}
