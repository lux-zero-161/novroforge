// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package wings

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"novroforge.xyz/models/db"
	"novroforge.xyz/modules/timeutil"

	"github.com/google/uuid"
)

const (
	BackupStatusPending   = "pending"
	BackupStatusCompleted = "completed"
	BackupStatusFailed    = "failed"

	BackupRestoreNone      = "none"
	BackupRestorePending   = "pending"
	BackupRestoreCompleted = "completed"
	BackupRestoreFailed    = "failed"

	maxBackupNameLength = 64
)

var (
	ErrInvalidBackupName          = errors.New("backup name must be between 1 and 64 characters")
	ErrBackupNotFound             = errors.New("Wings backup not found")
	ErrBackupNotReady             = errors.New("Wings backup is not ready for this operation")
	ErrInvalidBackupStatus        = errors.New("Wings backup status is invalid")
	ErrBackupCallbackUnauthorized = errors.New("Wings backup callback is unauthorized")
)

// WingsBackup stores backup metadata reported by the registered Wings node.
type WingsBackup struct {
	ID            int64              `xorm:"pk autoincr"`
	UUID          string             `xorm:"UNIQUE CHAR(36) NOT NULL"`
	ServerID      int64              `xorm:"INDEX NOT NULL"`
	Name          string             `xorm:"VARCHAR(64) NOT NULL"`
	Status        string             `xorm:"VARCHAR(16) NOT NULL DEFAULT 'pending'"`
	Checksum      string             `xorm:"VARCHAR(128)"`
	ChecksumType  string             `xorm:"VARCHAR(16)"`
	Size          int64              `xorm:"NOT NULL DEFAULT 0"`
	RestoreStatus string             `xorm:"VARCHAR(16) NOT NULL DEFAULT 'none'"`
	CreatedUnix   timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix   timeutil.TimeStamp `xorm:"updated"`
}

func init() {
	db.RegisterModel(new(WingsBackup))
}

func normalizeBackupName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxBackupNameLength {
		return "", ErrInvalidBackupName
	}
	return name, nil
}

func validateBackupStatus(successful bool, checksum, checksumType string, size int64) error {
	if size < 0 {
		return ErrInvalidBackupStatus
	}
	if !successful {
		return nil
	}
	if checksumType != "sha1" || len(checksum) != 40 {
		return ErrInvalidBackupStatus
	}
	if _, err := hex.DecodeString(checksum); err != nil {
		return ErrInvalidBackupStatus
	}
	return nil
}

// CreateBackupRecord stores the pending metadata before asking Wings to start work.
func CreateBackupRecord(ctx context.Context, serverID int64, name string) (*WingsBackup, error) {
	if serverID <= 0 {
		return nil, errors.New("invalid Wings server ID")
	}
	name, err := normalizeBackupName(name)
	if err != nil {
		return nil, err
	}
	if _, exists, err := GetServerByID(ctx, serverID); err != nil {
		return nil, err
	} else if !exists {
		return nil, fmt.Errorf("Wings server %d does not exist", serverID)
	}

	backup := &WingsBackup{
		UUID:          uuid.NewString(),
		ServerID:      serverID,
		Name:          name,
		Status:        BackupStatusPending,
		RestoreStatus: BackupRestoreNone,
	}
	if _, err := db.GetEngine(ctx).Insert(backup); err != nil {
		return nil, fmt.Errorf("create pending Wings backup: %w", err)
	}
	return backup, nil
}

// ListServerBackups returns the most recent backup records for one server.
func ListServerBackups(ctx context.Context, serverID int64) ([]*WingsBackup, error) {
	if serverID <= 0 {
		return nil, errors.New("invalid Wings server ID")
	}
	backups := make([]*WingsBackup, 0)
	if err := db.GetEngine(ctx).Where("server_id = ?", serverID).Desc("created_unix").Find(&backups); err != nil {
		return nil, fmt.Errorf("list Wings backups for server %d: %w", serverID, err)
	}
	return backups, nil
}

// GetBackupByUUID finds backup metadata by its canonical UUID.
func GetBackupByUUID(ctx context.Context, backupUUID string) (*WingsBackup, bool, error) {
	parsed, err := uuid.Parse(backupUUID)
	if err != nil || parsed.String() != strings.ToLower(backupUUID) {
		return nil, false, ErrBackupNotFound
	}
	var backup WingsBackup
	exists, err := db.GetEngine(ctx).Where("uuid = ?", parsed.String()).Get(&backup)
	if err != nil {
		return nil, false, fmt.Errorf("get Wings backup %s: %w", parsed, err)
	}
	return &backup, exists, nil
}

// MarkBackupRequestFailed records a synchronous failure to queue a backup.
func MarkBackupRequestFailed(ctx context.Context, backupID int64) error {
	affected, err := db.GetEngine(ctx).ID(backupID).Where("status = ?", BackupStatusPending).
		Cols("status").Update(&WingsBackup{Status: BackupStatusFailed})
	if err != nil {
		return fmt.Errorf("mark Wings backup request %d as failed: %w", backupID, err)
	}
	if affected != 1 {
		return fmt.Errorf("mark Wings backup request %d as failed: expected to update one pending row, updated %d", backupID, affected)
	}
	return nil
}

// ApplyBackupStatus applies one authenticated callback from Wings.
func ApplyBackupStatus(ctx context.Context, backupUUID string, successful bool, checksum, checksumType string, size int64) error {
	if err := validateBackupStatus(successful, checksum, checksumType, size); err != nil {
		return err
	}
	if !successful {
		checksum = ""
		checksumType = ""
		size = 0
	}
	backup, exists, err := GetBackupByUUID(ctx, backupUUID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrBackupNotFound
	}
	status := BackupStatusFailed
	if successful {
		status = BackupStatusCompleted
	}
	affected, err := db.GetEngine(ctx).ID(backup.ID).Where("status = ?", BackupStatusPending).Cols("status", "checksum", "checksum_type", "size").
		Update(&WingsBackup{Status: status, Checksum: checksum, ChecksumType: checksumType, Size: size})
	if err != nil {
		return fmt.Errorf("update Wings backup %s status: %w", backupUUID, err)
	}
	if affected == 1 {
		return nil
	}
	backup, exists, err = GetBackupByUUID(ctx, backupUUID)
	if err != nil {
		return err
	}
	if exists && backup.Status == status && backup.Checksum == checksum && backup.ChecksumType == checksumType && backup.Size == size {
		return nil
	}
	return ErrBackupNotReady
}

// AuthenticateBackupCallback verifies the node credentials for a backup UUID.
func AuthenticateBackupCallback(ctx context.Context, backupUUID, tokenID, token string) error {
	backup, exists, err := GetBackupByUUID(ctx, backupUUID)
	if errors.Is(err, ErrBackupNotFound) {
		return ErrBackupCallbackUnauthorized
	}
	if err != nil {
		return err
	}
	if !exists {
		return ErrBackupCallbackUnauthorized
	}
	server, exists, err := GetServerByID(ctx, backup.ServerID)
	if err != nil {
		return fmt.Errorf("load server for Wings backup callback: %w", err)
	}
	if !exists {
		return ErrBackupCallbackUnauthorized
	}
	node, exists, err := GetNodeByID(ctx, server.NodeID)
	if err != nil {
		return fmt.Errorf("load node for Wings backup callback: %w", err)
	}
	if !exists {
		return ErrBackupCallbackUnauthorized
	}
	matches, err := node.AuthenticateRemoteCallback(tokenID, token)
	if err != nil {
		return err
	}
	if !matches {
		return ErrBackupCallbackUnauthorized
	}
	return nil
}

// SetBackupRestoreStatus applies the result of one authenticated restore callback.
func SetBackupRestoreStatus(ctx context.Context, backupUUID string, successful bool) error {
	backup, exists, err := GetBackupByUUID(ctx, backupUUID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrBackupNotFound
	}
	status := BackupRestoreFailed
	if successful {
		status = BackupRestoreCompleted
	}
	affected, err := db.GetEngine(ctx).ID(backup.ID).Where("restore_status = ?", BackupRestorePending).
		Cols("restore_status").Update(&WingsBackup{RestoreStatus: status})
	if err != nil {
		return fmt.Errorf("update Wings backup %s restore status: %w", backupUUID, err)
	}
	if affected == 1 {
		return nil
	}
	if backup.RestoreStatus == status {
		return nil
	}
	return ErrBackupNotReady
}

// MarkBackupRestoring marks a completed backup before sending a restore request.
func MarkBackupRestoring(ctx context.Context, backupUUID string) (*WingsBackup, error) {
	backup, exists, err := GetBackupByUUID(ctx, backupUUID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrBackupNotFound
	}
	if backup.Status != BackupStatusCompleted || backup.RestoreStatus == BackupRestorePending {
		return nil, ErrBackupNotReady
	}
	affected, err := db.GetEngine(ctx).ID(backup.ID).Where("status = ? AND restore_status != ?", BackupStatusCompleted, BackupRestorePending).
		Cols("restore_status").Update(&WingsBackup{RestoreStatus: BackupRestorePending})
	if err != nil {
		return nil, fmt.Errorf("mark Wings backup %s as restoring: %w", backupUUID, err)
	}
	if affected != 1 {
		return nil, ErrBackupNotReady
	}
	backup.RestoreStatus = BackupRestorePending
	return backup, nil
}

// MarkBackupRestoreRequestFailed records a restore rejected before Wings accepted it.
func MarkBackupRestoreRequestFailed(ctx context.Context, backupUUID string) error {
	backup, exists, err := GetBackupByUUID(ctx, backupUUID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrBackupNotFound
	}
	affected, err := db.GetEngine(ctx).ID(backup.ID).Where("restore_status = ?", BackupRestorePending).
		Cols("restore_status").Update(&WingsBackup{RestoreStatus: BackupRestoreFailed})
	if err != nil {
		return fmt.Errorf("mark Wings backup %s restore request as failed: %w", backupUUID, err)
	}
	if affected == 1 || backup.RestoreStatus == BackupRestoreFailed {
		return nil
	}
	return ErrBackupNotReady
}

// DeleteBackupRecord removes backup metadata after its local archive is deleted.
func DeleteBackupRecord(ctx context.Context, backupID int64) error {
	affected, err := db.GetEngine(ctx).ID(backupID).Where("status != ? AND restore_status != ?", BackupStatusPending, BackupRestorePending).Delete(&WingsBackup{})
	if err != nil {
		return fmt.Errorf("delete Wings backup record %d: %w", backupID, err)
	}
	if affected != 1 {
		return ErrBackupNotReady
	}
	return nil
}
