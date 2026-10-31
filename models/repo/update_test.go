// Copyright 2025 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo_test

import (
	"testing"

	"novroforge.xyz/models/db"
	repo_model "novroforge.xyz/models/repo"
	"novroforge.xyz/models/unittest"
	user_model "novroforge.xyz/models/user"
	"novroforge.xyz/modules/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckCreateRepository(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	t.Run("Success", func(t *testing.T) {
		user := &user_model.User{MaxRepoCreation: 1}

		err := repo_model.CheckCreateRepository(db.DefaultContext, user, user, "testName")

		require.NoError(t, err)
	})

	t.Run("AdminIgnoresRepoLimit", func(t *testing.T) {
		user := &user_model.User{MaxRepoCreation: 0, IsAdmin: true}

		err := repo_model.CheckCreateRepository(db.DefaultContext, user, user, "testName")

		require.NoError(t, err)
	})

	t.Run("RepoLimitReached", func(t *testing.T) {
		user := &user_model.User{MaxRepoCreation: 0}

		err := repo_model.CheckCreateRepository(db.DefaultContext, user, user, "testName")

		require.ErrorIs(t, err, repo_model.ErrReachLimitOfRepo{})
	})

	t.Run("UnusableRepoName", func(t *testing.T) {
		user := &user_model.User{MaxRepoCreation: 1}

		err := repo_model.CheckCreateRepository(db.DefaultContext, user, user, "testName/")

		require.ErrorIs(t, err, db.ErrNameCharsNotAllowed{Name: "testName/"})
	})

	t.Run("RepoAlreadyExists", func(t *testing.T) {
		unittest.AssertExistsIf(t, true, &repo_model.Repository{Name: "repo1"})
		user := &user_model.User{MaxRepoCreation: 2}

		err := repo_model.CheckCreateRepository(db.DefaultContext, user, user, "repo1")

		require.ErrorIs(t, err, repo_model.ErrRepoAlreadyExist{Name: "repo1"})
	})

	t.Run("RepoDirAlreadyExists", func(t *testing.T) {
		db.DeleteByBean(db.DefaultContext, &repo_model.Repository{Name: "repo1"})
		user := &user_model.User{MaxRepoCreation: 2, Name: "user2"}

		exists, _ := util.IsExist(repo_model.RepoPath("user2", "repo1"))
		assert.True(t, exists)

		err := repo_model.CheckCreateRepository(db.DefaultContext, user, user, "repo1")

		require.ErrorIs(t, err, repo_model.ErrRepoAlreadyExist{Name: "repo1", Uname: "user2"})
	})
}
