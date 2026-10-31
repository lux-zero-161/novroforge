// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"testing"

	"novroforge.xyz/modules/setting"
	"novroforge.xyz/modules/test"
	"novroforge.xyz/routers"
	"novroforge.xyz/tests"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func getLinks(t *testing.T, url string) []*html.Node {
	req := NewRequest(t, "GET", url)
	resp := MakeRequest(t, req, http.StatusOK)

	htmlDoc := NewHTMLParser(t, resp.Body)
	links := htmlDoc.doc.Find("link[type=\"application/activity+json\"]").Nodes

	return links
}

func TestFederationBaseHead(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	defer test.MockVariableValue(&testWebRoutes, routers.NormalRoutes())()

	t.Run("Federation disabled", func(t *testing.T) {
		defer test.MockVariableValue(&setting.Federation.Enabled, false)()

		links := getLinks(t, "/user1")
		assert.Empty(t, links)
	})

	t.Run("Federation enabled", func(t *testing.T) {
		defer test.MockVariableValue(&setting.Federation.Enabled, true)()

		links := getLinks(t, "/user1")
		assert.Len(t, links, 1)
	})

	t.Run("Organization", func(t *testing.T) {
		defer test.MockVariableValue(&setting.Federation.Enabled, true)()

		links := getLinks(t, "/org3")
		assert.Empty(t, links)
	})
}
