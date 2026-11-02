// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package authz

import (
	"context"

	"novroforge.xyz/models/perm"
	repo_model "novroforge.xyz/models/repo"

	"xorm.io/builder"
)

// Implementation of [AuthorizationReducer] that does no authorization reduction, allowing normal access to all
// resources.
type AllAccessAuthorizationReducer struct{}

func (*AllAccessAuthorizationReducer) ReduceRepoAccess(ctx context.Context, repo *repo_model.Repository, accessMode perm.AccessMode) (perm.AccessMode, error) {
	return accessMode, nil
}

func (*AllAccessAuthorizationReducer) RepoReadAccessFilter() builder.Cond {
	return builder.NewCond() // invalid cond should be excluded and cause no filtering
}

func (*AllAccessAuthorizationReducer) AllowAdminOverride() bool {
	return true
}
