// Copyright 2024 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repository

import (
	"context"

	"novroforge.xyz/models/repo"
	"novroforge.xyz/models/user"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/services/federation"
)

func StarRepoAndSendLikeActivities(ctx context.Context, doer user.User, repoID int64, star bool) error {
	if err := repo.StarRepo(ctx, doer.ID, repoID, star); err != nil {
		return err
	}

	if star && setting.Federation.Enabled {
		if err := federation.SendLikeActivities(ctx, doer, repoID); err != nil {
			return err
		}
	}

	return nil
}
