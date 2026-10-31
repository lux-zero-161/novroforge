// Copyright The Forgejo Authors.
// SPDX-License-Identifier: MIT

package integration

import (
	"net/http"
	"testing"

	auth_model "novroforge.xyz/models/auth"
	"novroforge.xyz/modules/setting"
	"novroforge.xyz/modules/test"
	"novroforge.xyz/routers"
	v1 "novroforge.xyz/routers/api/forgejo/v1"
	"novroforge.xyz/tests"

	"github.com/stretchr/testify/assert"
)

func TestAPIForgejoVersion(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	t.Run("Version", func(t *testing.T) {
		req := NewRequest(t, "GET", "/api/forgejo/v1/version")
		resp := MakeRequest(t, req, http.StatusOK)

		var version v1.Version
		DecodeJSON(t, resp, &version)
		assert.Equal(t, "1.0.0", *version.Version)
	})

	t.Run("Version with Content-Type is json", func(t *testing.T) {
		req := NewRequest(t, "GET", "/api/forgejo/v1/version")
		resp := MakeRequest(t, req, http.StatusOK)

		assert.Equal(t, "application/json; charset=utf-8", resp.Header().Get("Content-Type"))
	})

	t.Run("Versions with REQUIRE_SIGNIN_VIEW enabled", func(t *testing.T) {
		defer test.MockVariableValue(&setting.Service.RequireSignInView, true)()
		defer test.MockVariableValue(&testWebRoutes, routers.NormalRoutes())()

		t.Run("Get forgejo version without auth", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()

			// GET api without auth
			req := NewRequest(t, "GET", "/api/forgejo/v1/version")
			MakeRequest(t, req, http.StatusForbidden)
		})

		t.Run("Get forgejo version without auth", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			username := "user1"
			session := loginUser(t, username)
			token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

			// GET api with auth
			req := NewRequest(t, "GET", "/api/forgejo/v1/version").AddTokenAuth(token)
			resp := MakeRequest(t, req, http.StatusOK)

			var version v1.Version
			DecodeJSON(t, resp, &version)
			assert.Equal(t, "1.0.0", *version.Version)
		})
	})
}
