package integration

import (
	"net/http"
	"testing"

	"novroforge.xyz/modules/setting"
	"novroforge.xyz/routers/web/healthcheck"
	"novroforge.xyz/tests"

	"github.com/stretchr/testify/assert"
)

func TestApiHeatlhCheck(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/api/healthz")
	resp := MakeRequest(t, req, http.StatusOK)
	assert.Contains(t, resp.Header().Values("Cache-Control"), "no-store")

	var status healthcheck.Response
	DecodeJSON(t, resp, &status)
	assert.Equal(t, healthcheck.Pass, status.Status)
	assert.Equal(t, setting.AppName, status.Description)
}
