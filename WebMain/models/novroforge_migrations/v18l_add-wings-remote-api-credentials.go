// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"novroforge.xyz/modules/timeutil"

	"novroforge/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add Wings remote API token IDs",
		Upgrade:     addWingsRemoteAPICredentials,
	})
}

func addWingsRemoteAPICredentials(x *xorm.Engine) error {
	type WingsNode struct {
		ID             int64              `xorm:"pk autoincr"`
		Name           string             `xorm:"UNIQUE NOT NULL VARCHAR(64)"`
		URL            string             `xorm:"TEXT NOT NULL"`
		EncryptedToken []byte             `xorm:"BLOB"`
		RemoteTokenID  string             `xorm:"VARCHAR(16)"`
		CreatedUnix    timeutil.TimeStamp `xorm:"created"`
		UpdatedUnix    timeutil.TimeStamp `xorm:"updated"`
	}

	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(WingsNode))
	return err
}
