// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package swagger

import (
	"novroforge.xyz/modules/forgefed"
	api "novroforge.xyz/modules/structs"
)

// ActivityPub
// swagger:response ActivityPub
type swaggerResponseActivityPub struct {
	// in:body
	Body api.ActivityPub `json:"body"`
}

// Outbox
// swagger:response Outbox
type swaggerResponseOutbox struct {
	// in:body
	Body forgefed.ForgeOutbox `json:"body"`
}
