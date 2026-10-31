// Copyright The NovroForge Authors.
// SPDX-License-Identifier: MIT

package integration

import (
	"net/http"
	"testing"

	"novroforge.xyz/tests"

	"github.com/stretchr/testify/assert"
)

func TestAPINovroForgeRoot(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/api/forgejo/v1")
	resp := MakeRequest(t, req, http.StatusNoContent)
	assert.Contains(t, resp.Header().Get("Link"), "/assets/novroforge/api.v1.yml")
}
