// Copyright 2025 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"testing"

	"novroforge.xyz/tests"

	"github.com/stretchr/testify/assert"
)

// TestUserHasNonExistentTheme verifies the UI falls back to the default theme when the
// user's DB entry has a theme that is no longer available.
func TestUserHasNonExistentTheme(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	// user40 has a "frog" theme in the DB
	user := loginUser(t, "user40")

	response := user.MakeRequest(t, NewRequest(t, "GET", "/user/settings/appearance"), http.StatusOK)
	page := NewHTMLParser(t, response.Body)

	dataTheme, dataThemeExists := page.Find("html").Attr("data-theme")
	assert.True(t, dataThemeExists)
	assert.Equal(t, "github-auto", dataTheme)

	assert.Equal(t, 0, page.Find("form[action='/user/settings/appearance/theme']").Length())
}
