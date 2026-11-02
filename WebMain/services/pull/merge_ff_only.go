// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package pull

import (
	repo_model "novroforge.xyz/models/repo"
	"novroforge.xyz/modules/git"
	"novroforge.xyz/modules/log"
)

// doMergeStyleFastForwardOnly merges the tracking into the current HEAD - which is assumed to be staging branch (equal to the pr.BaseBranch)
func doMergeStyleFastForwardOnly(ctx *mergeContext) error {
	cmd := git.NewCommand(ctx, "merge", "--ff-only").AddDynamicArguments(trackingBranch)
	if err := runMergeCommand(ctx, repo_model.MergeStyleFastForwardOnly, cmd); err != nil {
		log.Error("%-v Unable to merge tracking into base: %v", ctx.pr, err)
		return err
	}

	return nil
}
