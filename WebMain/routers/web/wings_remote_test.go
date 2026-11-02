// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package web

import "testing"

func TestParseWingsRemoteAuthorization(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		wantTokenID string
		wantToken   string
		wantOK      bool
	}{
		{name: "valid", header: "Bearer abcdefghijklmnop." + "secret", wantTokenID: "abcdefghijklmnop", wantToken: "secret", wantOK: true},
		{name: "case insensitive scheme", header: "bearer abcdefghijklmnop.secret", wantTokenID: "abcdefghijklmnop", wantToken: "secret", wantOK: true},
		{name: "missing scheme", header: "abcdefghijklmnop.secret"},
		{name: "missing separator", header: "Bearer abcdefghijklmnop"},
		{name: "empty token", header: "Bearer abcdefghijklmnop."},
		{name: "multiple separators", header: "Bearer abcdefghijklmnop.secret.more"},
		{name: "extra whitespace", header: "Bearer  abcdefghijklmnop.secret"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotID, gotToken, gotOK := parseWingsRemoteAuthorization(test.header)
			if gotID != test.wantTokenID || gotToken != test.wantToken || gotOK != test.wantOK {
				t.Errorf("parseWingsRemoteAuthorization() = (%q, %q, %t), want (%q, %q, %t)", gotID, gotToken, gotOK, test.wantTokenID, test.wantToken, test.wantOK)
			}
		})
	}
}
