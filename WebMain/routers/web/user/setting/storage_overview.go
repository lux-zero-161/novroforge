// Copyright 2024 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package setting

import (
	"novroforge.xyz/modules/base"
	"novroforge.xyz/routers/web/shared"
	"novroforge.xyz/services/context"
)

const (
	tplSettingsStorageOverview base.TplName = "user/settings/storage_overview"
)

// StorageOverview render a size overview of the user, as well as relevant
// quota limits of the instance.
func StorageOverview(ctx *context.Context) {
	shared.StorageOverview(ctx, ctx.Doer.ID, tplSettingsStorageOverview)
}
