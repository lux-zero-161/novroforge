// Copyright 2026 NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package automerge

import (
	"testing"

	"novroforge.xyz/models/db"
	issues_model "novroforge.xyz/models/issues"
	"novroforge.xyz/models/perm"
	access_model "novroforge.xyz/models/perm/access"
	pull_model "novroforge.xyz/models/pull"
	"novroforge.xyz/models/unit"
	"novroforge.xyz/models/unittest"
	user_model "novroforge.xyz/models/user"
	"novroforge.xyz/modules/util"

	"github.com/stretchr/testify/require"
)

func TestRemoveScheduledAutoMerge(t *testing.T) {
	defer unittest.OverrideFixtures("services/automerge/fixtures/TestRemoveScheduledAutoMerge")()
	require.NoError(t, unittest.PrepareTestDatabase())

	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	user5 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 5})
	pull1 := unittest.AssertExistsAndLoadBean(t, &issues_model.PullRequest{ID: 1})
	pull2 := unittest.AssertExistsAndLoadBean(t, &issues_model.PullRequest{ID: 2})

	t.Run("No automerge", func(t *testing.T) {
		err := RemoveScheduledAutoMerge(t.Context(), user5, pull2, access_model.Permission{})
		require.ErrorIs(t, err, db.ErrNotExist{Resource: "auto_merge", ID: 2})
	})

	t.Run("No permission", func(t *testing.T) {
		err := RemoveScheduledAutoMerge(t.Context(), user5, pull1, access_model.Permission{})
		require.ErrorIs(t, err, util.ErrPermissionDenied)

		err = RemoveScheduledAutoMerge(t.Context(), user5, pull1, access_model.Permission{UnitsMode: map[unit.Type]perm.AccessMode{
			unit.TypePullRequests: perm.AccessModeRead,
		}})
		require.ErrorIs(t, err, util.ErrPermissionDenied)
	})

	t.Run("Normal", func(t *testing.T) {
		err := RemoveScheduledAutoMerge(t.Context(), user2, pull1, access_model.Permission{UnitsMode: map[unit.Type]perm.AccessMode{
			unit.TypePullRequests: perm.AccessModeWrite,
		}})
		require.NoError(t, err)

		unittest.AssertExistsIf(t, false, &pull_model.AutoMerge{PullID: pull1.ID})
		unittest.AssertExistsIf(t, true, &issues_model.Comment{IssueID: pull1.IssueID, Type: issues_model.CommentTypePRUnScheduledToAutoMerge})
	})
}
