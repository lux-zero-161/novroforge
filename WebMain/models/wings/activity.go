// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package wings

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"novroforge.xyz/models/db"
	"novroforge.xyz/modules/timeutil"
)

const (
	ActivityPower             = "power"
	ActivityCommand           = "command"
	ActivityInstall           = "install"
	ActivityReinstall         = "reinstall"
	ActivitySync              = "sync"
	ActivityDelete            = "delete"
	ActivityRegister          = "register"
	ActivityRename            = "rename"
	ActivityRemove            = "remove"
	ActivityGrantAccess       = "access_grant"
	ActivityRevokeAccess      = "access_revoke"
	ActivityNodeDelete        = "node_delete"
	ActivityNodeTokenRotate   = "node_token_rotate"
	ActivityNodeRemoteCreds   = "node_remote_credentials"
	ActivityFileEdit          = "file_edit"
	ActivityFileCreateDir     = "file_create_directory"
	ActivityFileRename        = "file_rename"
	ActivityFileDelete        = "file_delete"
	ActivityFileCopy          = "file_copy"
	ActivityFileCompress      = "file_compress"
	ActivityFileDecompress    = "file_decompress"
	ActivityFileChmod         = "file_chmod"
	ActivityFileUpload        = "file_upload"
	ActivityBackupCreate      = "backup_create"
	ActivityBackupRestore     = "backup_restore"
	ActivityBackupDelete      = "backup_delete"
	activityEventMaxRunes     = 64
	activityDetailMaxRunes    = 255
	activityActorNameMaxRunes = 255
	defaultActivityLimit      = 20
	maxActivityLimit          = 100
)

var ErrInvalidActivity = errors.New("invalid Wings activity entry")

// WingsActivity records a user action that affected a Wings node or server.
type WingsActivity struct {
	ID          int64              `xorm:"pk autoincr"`
	NodeID      int64              `xorm:"INDEX NOT NULL"`
	ServerID    int64              `xorm:"INDEX NOT NULL DEFAULT 0"`
	ActorID     int64              `xorm:"INDEX NOT NULL"`
	ActorName   string             `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
	Event       string             `xorm:"VARCHAR(64) NOT NULL"`
	Detail      string             `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
	CreatedUnix timeutil.TimeStamp `xorm:"created"`
}

func init() {
	db.RegisterModel(new(WingsActivity))
}

// EventLabelKey returns the locale key used to render the event name.
func (activity *WingsActivity) EventLabelKey() string {
	return "wings.activity_event_" + activity.Event
}

// RecordActivity stores one Wings activity entry. Detail is truncated to the
// column limit; tokens and console command text must never be passed here.
func RecordActivity(ctx context.Context, nodeID, serverID, actorID int64, actorName, event, detail string) error {
	if nodeID <= 0 || serverID < 0 || actorID <= 0 {
		return ErrInvalidActivity
	}
	event = strings.TrimSpace(event)
	if event == "" || utf8.RuneCountInString(event) > activityEventMaxRunes {
		return ErrInvalidActivity
	}

	activity := &WingsActivity{
		NodeID:    nodeID,
		ServerID:  serverID,
		ActorID:   actorID,
		ActorName: truncateRunes(strings.TrimSpace(actorName), activityActorNameMaxRunes),
		Event:     event,
		Detail:    truncateRunes(strings.TrimSpace(detail), activityDetailMaxRunes),
	}
	if _, err := db.GetEngine(ctx).Insert(activity); err != nil {
		return fmt.Errorf("record Wings activity %q: %w", event, err)
	}
	return nil
}

// ListServerActivity returns the newest entries for one registered server.
func ListServerActivity(ctx context.Context, serverID int64, limit int) ([]*WingsActivity, error) {
	if serverID <= 0 {
		return nil, ErrInvalidActivity
	}
	activities := make([]*WingsActivity, 0, clampActivityLimit(limit))
	if err := db.GetEngine(ctx).
		Where("server_id = ?", serverID).
		OrderBy("id DESC").
		Limit(clampActivityLimit(limit)).
		Find(&activities); err != nil {
		return nil, fmt.Errorf("list Wings activity for server %d: %w", serverID, err)
	}
	return activities, nil
}

// ListActivity returns the newest entries across all nodes and servers.
func ListActivity(ctx context.Context, limit int) ([]*WingsActivity, error) {
	activities := make([]*WingsActivity, 0, clampActivityLimit(limit))
	if err := db.GetEngine(ctx).
		OrderBy("id DESC").
		Limit(clampActivityLimit(limit)).
		Find(&activities); err != nil {
		return nil, fmt.Errorf("list Wings activity: %w", err)
	}
	return activities, nil
}

func clampActivityLimit(limit int) int {
	if limit <= 0 {
		return defaultActivityLimit
	}
	if limit > maxActivityLimit {
		return maxActivityLimit
	}
	return limit
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 || value == "" {
		return ""
	}
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit])
}
