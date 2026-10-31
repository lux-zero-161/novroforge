// Copyright 2024 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package novroforge_migrations_legacy

import "code.novroforge.xyz/xorm/xorm"

func AddSSHKeypairToPushMirror(x *xorm.Engine) error {
	type PushMirror struct {
		ID         int64  `xorm:"pk autoincr"`
		PublicKey  string `xorm:"VARCHAR(100)"`
		PrivateKey []byte `xorm:"BLOB"`
	}

	return x.Sync(&PushMirror{})
}
