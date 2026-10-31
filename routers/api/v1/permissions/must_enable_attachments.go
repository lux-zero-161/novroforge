// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"novroforge.xyz/modules/setting"
)

func MustEnableAttachments(ctx Context) {
	if !setting.Attachment.Enabled {
		ctx.NotFound()
		return
	}
}
