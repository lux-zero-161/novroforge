// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package actions

import (
	"testing"

	"novroforge.xyz/models/db"
	"novroforge.xyz/models/unittest"
	"novroforge.xyz/modules/optional"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLatestRunnerToken(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	token := unittest.AssertExistsAndLoadBean(t, &ActionRunnerToken{ID: 3})
	expectedToken, err := GetLatestRunnerToken(db.DefaultContext, optional.Some[int64](1), optional.None[int64]())
	require.NoError(t, err)
	assert.Equal(t, expectedToken, token)
}

func TestNewRunnerToken(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	token, err := NewRunnerToken(db.DefaultContext, optional.Some[int64](1), optional.None[int64]())
	require.NoError(t, err)
	expectedToken, err := GetLatestRunnerToken(db.DefaultContext, optional.Some[int64](1), optional.None[int64]())
	require.NoError(t, err)
	assert.Equal(t, expectedToken, token)
}

func TestUpdateRunnerToken(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	token := unittest.AssertExistsAndLoadBean(t, &ActionRunnerToken{ID: 3})
	token.IsActive = true
	require.NoError(t, UpdateRunnerToken(db.DefaultContext, token, "is_active"))
	expectedToken, err := GetLatestRunnerToken(db.DefaultContext, optional.Some[int64](1), optional.None[int64]())
	require.NoError(t, err)
	assert.Equal(t, expectedToken, token)
}
