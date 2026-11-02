// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package wings

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeServerName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{name: "trims surrounding whitespace", input: "  Survival  ", want: "Survival"},
		{name: "accepts 64 characters", input: strings.Repeat("a", 64), want: strings.Repeat("a", 64)},
		{name: "counts Unicode code points", input: strings.Repeat("界", 64), want: strings.Repeat("界", 64)},
		{name: "rejects empty", wantErr: ErrInvalidServerName},
		{name: "rejects whitespace only", input: " \t ", wantErr: ErrInvalidServerName},
		{name: "rejects more than 64 characters", input: strings.Repeat("a", 65), wantErr: ErrInvalidServerName},
		{name: "rejects more than 64 Unicode code points", input: strings.Repeat("界", 65), wantErr: ErrInvalidServerName},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeServerName(test.input)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("normalizeServerName() error = %v, want %v", err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("normalizeServerName() = %q, want %q", got, test.want)
			}
		})
	}
}
