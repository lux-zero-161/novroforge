// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package misc

import (
	"testing"

	"novroforge.xyz/modules/setting"
	"novroforge.xyz/modules/test"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRobotsTxt(t *testing.T) {
	defer test.MockVariableValue(&setting.AppDocsVer, func() string { return "v0.0" })()

	robotsTxt := string(defaultRobotsTxt())
	assert.Contains(t, robotsTxt, "https://novroforge.xyz/docs/v0.0/admin/advanced/search-engines/") // derived from setting.AppDocsVer
	assert.Equal(t, robotsTxt, string(defaultRobotsTxt()), "subsequent call still works")
}
