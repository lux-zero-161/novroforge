// Copyright 2024 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forgejo_migrations_legacy

import "code.forgejo.org/xorm/xorm"

func AddPurposeToNovroForgeAuthToken(x *xorm.Engine) error {
	type NovroForgeAuthToken struct {
		ID      int64  `xorm:"pk autoincr"`
		Purpose string `xorm:"NOT NULL DEFAULT 'long_term_authorization'"`
	}
	if err := x.Sync(new(NovroForgeAuthToken)); err != nil {
		return err
	}

	_, err := x.Exec("UPDATE `forgejo_auth_token` SET purpose = 'long_term_authorization' WHERE purpose = ''")
	return err
}
