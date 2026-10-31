// Copyright 2025 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations_legacy

import "code.novroforge.xyz/xorm/xorm"

func AddNotifyEmailToActionRun(x *xorm.Engine) error {
	type ActionRun struct {
		ID          int64
		NotifyEmail bool
	}
	return x.Sync(new(ActionRun))
}
