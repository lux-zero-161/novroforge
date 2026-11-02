// Copyright 2025 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package stats

import (
	"context"

	"novroforge.xyz/modules/optional"
	"novroforge.xyz/modules/timeutil"
)

// Queue a recalculation of the stats on a `Milestone` for a given milestone by its ID
func QueueRecalcMilestoneByID(ctx context.Context, labelID int64) {
	safePush(ctx, recalcRequest{
		RecalcType: MilestoneByMilestoneID,
		ObjectID:   labelID,
	})
}

func QueueRecalcMilestoneByIDWithDate(ctx context.Context, labelID int64, updateTimestamp timeutil.TimeStamp) {
	safePush(ctx, recalcRequest{
		RecalcType:      MilestoneByMilestoneID,
		ObjectID:        labelID,
		UpdateTimestamp: optional.Some(updateTimestamp),
	})
}
