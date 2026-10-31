// Copyright 2025 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"testing"

	auth_model "novroforge.xyz/models/auth"
	"novroforge.xyz/models/forgefed"
	"novroforge.xyz/models/unittest"
	"novroforge.xyz/models/user"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/modules/structs"
	"novroforge.xyz/modules/test"
	"novroforge.xyz/routers"
	"novroforge.xyz/tests"

	"github.com/stretchr/testify/assert"
)

func TestActivityPubFollowFederated(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	defer test.MockVariableValue(&setting.Federation.Enabled, true)()
	defer test.MockVariableValue(&setting.Federation.InsecureAllowInvalidHosts, true)()
	defer test.MockVariableValue(&testWebRoutes, routers.NormalRoutes())()

	mock := test.NewFederationServerMock()
	federatedSrv := mock.DistantServer(t)
	defer federatedSrv.Close()

	localUser10Name := "user10"
	localSession10 := loginUser(t, localUser10Name)
	localSecssion10Token := getTokenForLoggedInUser(t, localSession10, auth_model.AccessTokenScopeWriteUser)

	distantURL := federatedSrv.URL
	distantUser15URL := fmt.Sprintf("%s/api/v1/activitypub/user-id/15", distantURL)

	// local user follow distant
	req := NewRequestWithJSON(t, "POST",
		"/api/v1/user/activitypub/follow",
		&structs.APRemoteFollowOption{
			Target: distantUser15URL,
		}).
		AddTokenAuth(localSecssion10Token)
	MakeRequest(t, req, http.StatusNoContent)

	// check: federated actors now exist local
	federationHost := unittest.AssertExistsAndLoadBean(t, &forgefed.FederationHost{HostFqdn: "127.0.0.1"})
	unittest.AssertExistsAndLoadBean(t, &user.FederatedUser{ExternalID: "15", FederationHostID: federationHost.ID})

	// check: follow request arrived at distant
	assert.Contains(t, mock.LastPost, "\"object\":\"http://DISTANT_FEDERATION_HOST/api/v1/activitypub/user-id/15\"")
}
