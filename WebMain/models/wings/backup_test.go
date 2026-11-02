// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package wings

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeBackupName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "trimmed", input: "  nightly  ", want: "nightly"},
		{name: "empty", input: "  ", wantErr: true},
		{name: "too long", input: strings.Repeat("x", maxBackupNameLength+1), wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeBackupName(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("normalizeBackupName() error = %v, wantErr %t", err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("normalizeBackupName() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestValidateBackupStatus(t *testing.T) {
	tests := []struct {
		name         string
		successful   bool
		checksum     string
		checksumType string
		size         int64
		wantErr      error
	}{
		{name: "valid success", successful: true, checksum: strings.Repeat("a", 40), checksumType: "sha1", size: 10},
		{name: "failure without details", successful: false},
		{name: "negative size", successful: false, size: -1, wantErr: ErrInvalidBackupStatus},
		{name: "unsupported checksum", successful: true, checksum: strings.Repeat("a", 40), checksumType: "sha256", wantErr: ErrInvalidBackupStatus},
		{name: "malformed checksum", successful: true, checksum: strings.Repeat("z", 40), checksumType: "sha1", wantErr: ErrInvalidBackupStatus},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateBackupStatus(test.successful, test.checksum, test.checksumType, test.size)
			if !errors.Is(err, test.wantErr) {
				t.Errorf("validateBackupStatus() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}
