// Copyright 2026 The Forgejo Authors
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"testing"

	"novroforge.xyz/tests"

	"github.com/stretchr/testify/assert"
)

func TestFooterContent(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	user := loginUser(t, "user2")
	page := NewHTMLParser(t, user.MakeRequest(t, NewRequest(t, "GET", "/explore/organizations"), http.StatusOK).Body)
	footer := page.Find("footer").Text()

	assert.Contains(t, footer, "Copyright © 2026 NovroForge")
	assert.NotContains(t, footer, "Powered by Forgejo")
	assert.NotContains(t, footer, "Version")
	assert.NotContains(t, footer, "Template")
	assert.NotContains(t, footer, "Licenses")
}
