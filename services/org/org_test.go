// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package org

import (
	"testing"

	"novroforge.xyz/models"
	"novroforge.xyz/models/actions"
	"novroforge.xyz/models/db"
	"novroforge.xyz/models/organization"
	"novroforge.xyz/models/unittest"
	user_model "novroforge.xyz/models/user"
	"novroforge.xyz/modules/optional"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}

func TestDeleteOrganization(t *testing.T) {
	defer unittest.OverrideFixtures("services/org/TestDeleteOrganization")()
	require.NoError(t, unittest.PrepareTestDatabase())

	org := unittest.AssertExistsAndLoadBean(t, &organization.Organization{ID: 6})
	require.NoError(t, DeleteOrganization(db.DefaultContext, org, false))
	unittest.AssertNotExistsBean(t, &organization.Organization{ID: 6})
	unittest.AssertNotExistsBean(t, &organization.OrgUser{OrgID: 6})
	unittest.AssertNotExistsBean(t, &organization.Team{OrgID: 6})
	unittest.AssertNotExistsBean(t, &actions.ActionRunnerToken{OwnerID: optional.Some[int64](6)})
	unittest.AssertNotExistsBean(t, &user_model.Follow{FollowID: 6})
	unittest.AssertNotExistsBean(t, &user_model.BlockedUser{UserID: 6})

	org = unittest.AssertExistsAndLoadBean(t, &organization.Organization{ID: 3})
	err := DeleteOrganization(db.DefaultContext, org, false)
	require.Error(t, err)
	assert.True(t, models.IsErrUserOwnRepos(err))

	user := unittest.AssertExistsAndLoadBean(t, &organization.Organization{ID: 5})
	require.Error(t, DeleteOrganization(db.DefaultContext, user, false))
	unittest.CheckConsistencyFor(t, &user_model.User{}, &organization.Team{})

	assert.Zero(t, unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1001}).NumFollowing)
}
