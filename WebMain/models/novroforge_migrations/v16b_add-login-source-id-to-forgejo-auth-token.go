// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"novroforge.xyz/modules/optional"

	"novroforge/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add login_source_id column to novroforge_auth_token",
		Upgrade:     addLoginSourceIDToNovroForgeAuthToken,
	})
}

func addLoginSourceIDToNovroForgeAuthToken(x *xorm.Engine) error {
	type NovroForgeAuthToken struct {
		LoginSourceID optional.Option[int64] `xorm:"INDEX REFERENCES(login_source, id)"`
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(NovroForgeAuthToken))
	return err
}
