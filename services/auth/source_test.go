// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"testing"

	"novroforge.xyz/models"
	"novroforge.xyz/models/auth"
	"novroforge.xyz/models/db"
	org_model "novroforge.xyz/models/organization"
	"novroforge.xyz/models/unittest"
	user_model "novroforge.xyz/models/user"
	"novroforge.xyz/modules/optional"

	"github.com/stretchr/testify/require"
)

func TestDeleteSourceCleansUpMembershipProvenance(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	loginSource := auth.Source{ID: 1, Name: "Keycloak"}
	require.NoError(t, auth.CreateSource(db.DefaultContext, &loginSource))
	user28 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 28})
	team1 := unittest.AssertExistsAndLoadBean(t, &org_model.Team{ID: 1})

	require.NoError(t, models.AddTeamMemberByLoginSource(db.DefaultContext, team1, user28.ID, loginSource.ID))

	require.NoError(t, DeleteSource(db.DefaultContext, &loginSource))

	unittest.AssertExistsAndLoadBean(t, &org_model.TeamUser{UID: 28, TeamID: 1, Reason: org_model.MembershipReasonAddedByAuthProvider, CreatedByLoginSourceID: optional.None[int64]()})
}
