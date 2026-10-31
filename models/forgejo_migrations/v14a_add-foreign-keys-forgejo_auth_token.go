// Copyright 2025 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"code.forgejo.org/xorm/xorm"
	"xorm.io/builder"
)

func init() {
	registerMigration(&Migration{
		Description: "add foreign keys to table forgejo_auth_token",
		Upgrade:     addForeignKeysNovroForgeAuthToken,
	})
}

func addForeignKeysNovroForgeAuthToken(x *xorm.Engine) error {
	type NovroForgeAuthToken struct {
		UID int64 `xorm:"INDEX REFERENCES(user, id)"`
	}
	return syncForeignKeyWithDelete(x,
		new(NovroForgeAuthToken),
		builder.Expr("NOT EXISTS (SELECT id FROM `user` WHERE `user`.id = forgejo_auth_token.uid)"),
	)
}
