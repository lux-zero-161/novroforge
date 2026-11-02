// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package user

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"novroforge.xyz/modules/setting"
)

func TestWingsConsoleOriginMatchesConfiguredPanelURLBehindProxy(t *testing.T) {
	previousURL := setting.AppURL
	setting.AppURL = "https://panel.example.test/panel/"
	t.Cleanup(func() {
		setting.AppURL = previousURL
	})

	tests := []struct {
		name   string
		origin string
		allow  bool
	}{
		{name: "configured public origin", origin: "https://panel.example.test", allow: true},
		{name: "different host", origin: "https://attacker.example.test", allow: false},
		{name: "different scheme", origin: "http://panel.example.test", allow: false},
		{name: "unexpected path", origin: "https://panel.example.test/panel", allow: false},
		{name: "query string", origin: "https://panel.example.test/?token=secret", allow: false},
		{name: "fragment", origin: "https://panel.example.test/#fragment", allow: false},
		{name: "missing origin", origin: "", allow: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "http://internal-proxy:8080/user/servers/1/console/ws", nil)
			request.Header.Set("Origin", test.origin)
			if got := wingsConsoleUpgrader.CheckOrigin(request); got != test.allow {
				t.Errorf("CheckOrigin(%q) = %t, want %t", test.origin, got, test.allow)
			}
		})
	}

	origin, err := wingsConsolePanelOrigin()
	if err != nil {
		t.Fatal(err)
	}
	if origin != (&url.URL{Scheme: "https", Host: "panel.example.test"}).String() {
		t.Fatalf("panel origin = %q", origin)
	}
}

func TestWingsConsoleConnectionLimit(t *testing.T) {
	serverID := time.Now().UnixNano()
	userID := serverID + 1
	releases := make([]func(), 0, maxWingsConsoleConnectionsPerUser)

	for range maxWingsConsoleConnectionsPerUser {
		release, acquired := acquireWingsConsoleConnection(serverID, userID)
		if !acquired {
			t.Fatal("expected console connection slot to be available")
		}
		releases = append(releases, release)
	}
	if release, acquired := acquireWingsConsoleConnection(serverID, userID); acquired {
		release()
		t.Fatal("expected per-user console connection limit to be enforced")
	}

	releases[0]()
	releases[0]()
	release, acquired := acquireWingsConsoleConnection(serverID, userID)
	if !acquired {
		t.Fatal("expected released console connection slot to be reusable")
	}
	release()
	for _, release := range releases[1:] {
		release()
	}
}
