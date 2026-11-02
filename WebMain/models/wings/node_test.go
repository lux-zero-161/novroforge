// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package wings

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateNodeToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  error
	}{
		{name: "valid", token: "visible-token_123"},
		{name: "empty", want: ErrInvalidNodeToken},
		{name: "leading whitespace", token: " token", want: ErrInvalidNodeToken},
		{name: "non-visible character", token: "token\n", want: ErrInvalidNodeToken},
		{name: "too long", token: strings.Repeat("a", 4097), want: ErrInvalidNodeToken},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateNodeToken(test.token)
			if !errors.Is(err, test.want) {
				t.Errorf("validateNodeToken() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestValidateRemoteCredentials(t *testing.T) {
	tests := []struct {
		name    string
		tokenID string
		token   string
		wantErr bool
	}{
		{name: "valid", tokenID: "abcdefghijklmnop", token: strings.Repeat("a", 64)},
		{name: "invalid token id length", tokenID: "short", token: strings.Repeat("a", 64), wantErr: true},
		{name: "invalid token id character", tokenID: "abcdefghijklmno.", token: strings.Repeat("a", 64), wantErr: true},
		{name: "short token", tokenID: "abcdefghijklmnop", token: "secret", wantErr: true},
		{name: "separator in token", tokenID: "abcdefghijklmnop", token: strings.Repeat("a", 32) + "." + strings.Repeat("b", 31), wantErr: true},
		{name: "non-visible token character", tokenID: "abcdefghijklmnop", token: strings.Repeat("a", 63) + "\n", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateRemoteCredentials(test.tokenID, test.token)
			if (err != nil) != test.wantErr {
				t.Errorf("validateRemoteCredentials() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
