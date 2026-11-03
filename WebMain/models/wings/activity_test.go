// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package wings

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestTruncateRunes(t *testing.T) {
	tests := []struct {
		name  string
		value string
		limit int
		want  string
	}{
		{name: "empty", value: "", limit: 10, want: ""},
		{name: "zero limit", value: "abc", limit: 0, want: ""},
		{name: "negative limit", value: "abc", limit: -1, want: ""},
		{name: "fits", value: "abc", limit: 10, want: "abc"},
		{name: "exact", value: "abc", limit: 3, want: "abc"},
		{name: "truncates ascii", value: "abcdef", limit: 3, want: "abc"},
		{name: "truncates runes not bytes", value: "a😀b😀c", limit: 3, want: "a😀b"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := truncateRunes(test.value, test.limit)
			if got != test.want {
				t.Errorf("truncateRunes(%q, %d) = %q, want %q", test.value, test.limit, got, test.want)
			}
		})
	}
}

func TestClampActivityLimit(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{input: 0, want: defaultActivityLimit},
		{input: -5, want: defaultActivityLimit},
		{input: 1, want: 1},
		{input: 20, want: 20},
		{input: maxActivityLimit, want: maxActivityLimit},
		{input: maxActivityLimit + 1, want: maxActivityLimit},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			got := clampActivityLimit(test.input)
			if got != test.want {
				t.Errorf("clampActivityLimit(%d) = %d, want %d", test.input, got, test.want)
			}
		})
	}
}

func TestEventLabelKey(t *testing.T) {
	activity := &WingsActivity{Event: ActivityPower}
	if got, want := activity.EventLabelKey(), "wings.activity_event_power"; got != want {
		t.Errorf("EventLabelKey() = %q, want %q", got, want)
	}
}

func TestRecordActivityValidation(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name      string
		nodeID    int64
		serverID  int64
		actorID   int64
		actorName string
		event     string
		detail    string
	}{
		{name: "zero node", nodeID: 0, serverID: 1, actorID: 1, event: ActivityPower},
		{name: "negative node", nodeID: -1, serverID: 1, actorID: 1, event: ActivityPower},
		{name: "negative server", nodeID: 1, serverID: -1, actorID: 1, event: ActivityPower},
		{name: "zero actor", nodeID: 1, serverID: 1, actorID: 0, event: ActivityPower},
		{name: "empty event", nodeID: 1, serverID: 1, actorID: 1, event: "   "},
		{name: "event too long", nodeID: 1, serverID: 1, actorID: 1, event: strings.Repeat("a", activityEventMaxRunes+1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := RecordActivity(ctx, test.nodeID, test.serverID, test.actorID, test.actorName, test.event, test.detail)
			if !errors.Is(err, ErrInvalidActivity) {
				t.Errorf("RecordActivity() error = %v, want ErrInvalidActivity", err)
			}
		})
	}
}

func TestRecordActivityTruncatesFields(t *testing.T) {
	// Only validates that long inputs are rejected at the event boundary and
	// that rune truncation helpers behave; full persistence needs a database.
	if got := truncateRunes(strings.Repeat("a", 300), activityDetailMaxRunes); len([]rune(got)) != activityDetailMaxRunes {
		t.Errorf("detail truncation length = %d, want %d", len([]rune(got)), activityDetailMaxRunes)
	}
	if got := truncateRunes(strings.Repeat("b", 300), activityActorNameMaxRunes); len([]rune(got)) != activityActorNameMaxRunes {
		t.Errorf("actor truncation length = %d, want %d", len([]rune(got)), activityActorNameMaxRunes)
	}
}

// keep package import surface minimal
