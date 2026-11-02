// Copyright 2025 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations_legacy

import (
	"novroforge/xorm/xorm"
)

// see https://codeberg.org/novroforge/novroforge/issues/8373
func NoopAddIndexToActionRunStopped(x *xorm.Engine) error {
	return nil
}
