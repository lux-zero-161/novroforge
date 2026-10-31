// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"net/http"
	"testing"

	"novroforge.xyz/modules/setting"
	api "novroforge.xyz/modules/structs"
	"novroforge.xyz/modules/test"
	"novroforge.xyz/routers"
	"novroforge.xyz/tests"

	"github.com/stretchr/testify/assert"
)

func TestNodeinfo(t *testing.T) {
	defer test.MockVariableValue(&setting.Federation.Enabled, true)()
	defer test.MockVariableValue(&testWebRoutes, routers.NormalRoutes())()
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/api/v1/nodeinfo")
	resp := MakeRequest(t, req, http.StatusOK)
	VerifyJSONSchema(t, resp, "nodeinfo_2.1.json")

	var nodeinfo api.NodeInfo
	DecodeJSON(t, resp, &nodeinfo)
	assert.True(t, nodeinfo.OpenRegistrations)
	assert.Equal(t, "novroforge", nodeinfo.Software.Name)
	assert.Equal(t, 31, nodeinfo.Usage.Users.Total)
	assert.Equal(t, 23, nodeinfo.Usage.LocalPosts)
	assert.Equal(t, 4, nodeinfo.Usage.LocalComments)
}
