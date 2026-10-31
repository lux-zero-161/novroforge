// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package models

import (
	"testing"

	activities_model "novroforge.xyz/models/activities"
	"novroforge.xyz/models/organization"
	repo_model "novroforge.xyz/models/repo"
	"novroforge.xyz/models/unittest"
	user_model "novroforge.xyz/models/user"

	"github.com/stretchr/testify/require"
)

// TestFixturesAreConsistent assert that test fixtures are consistent
func TestFixturesAreConsistent(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	unittest.CheckConsistencyFor(t,
		&user_model.User{},
		&repo_model.Repository{},
		&organization.Team{},
		&activities_model.Action{})
}

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
