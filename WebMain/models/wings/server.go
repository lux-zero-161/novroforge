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
	wings_service "novroforge.xyz/services/wings"

	"github.com/google/uuid"
)

var (
	ErrInvalidServerName = errors.New("server name must be between 1 and 64 characters")
	ErrInvalidServerUUID = errors.New("server UUID is invalid")
	ErrNodeHasServers    = errors.New("node still has registered servers")
)

// WingsServer is a WebMain registration for a server provisioned on Wings.
type WingsServer struct {
	ID             int64
	NodeID         int64              `xorm:"UNIQUE(node_uuid) INDEX NOT NULL"`
	UUID           string             `xorm:"UNIQUE(node_uuid) CHAR(36) NOT NULL"`
	EggID          int64              `xorm:"INDEX NOT NULL DEFAULT 0"`
	AllocationID   int64              `xorm:"INDEX NOT NULL DEFAULT 0"`
	Name           string             `xorm:"VARCHAR(64) NOT NULL"`
	CreatedUnix    timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix    timeutil.TimeStamp `xorm:"updated"`
}

// ServerListItem contains only safe display data for administrator pages.
type ServerListItem struct {
	ID       int64
	NodeID   int64
	NodeName string
	Name     string
	UUID     string
	Accesses []ServerUserListItem
}

func init() {
	db.RegisterModel(new(WingsServer))
}

// CreateServer registers an existing Wings server after verifying it on its node.
func CreateServer(ctx context.Context, nodeID, eggID, allocationID int64, name, serverUUID string) (int64, error) {
	name, err := normalizeServerName(name)
	if err != nil {
		return 0, ErrInvalidServerName
	}
	parsedUUID, err := uuid.Parse(strings.TrimSpace(serverUUID))
	if err != nil {
		return 0, ErrInvalidServerUUID
	}
	node, exists, err := GetNodeByID(ctx, nodeID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, fmt.Errorf("Wings node %d does not exist", nodeID)
	}
	client, err := node.GetClient()
	if err != nil {
		return 0, err
	}
	if _, err := client.GetServer(ctx, parsedUUID.String()); err != nil {
		return 0, fmt.Errorf("verify server on Wings node: %w", err)
	}

	server := &WingsServer{
		NodeID:       nodeID,
		EggID:        eggID,
		AllocationID: allocationID,
		UUID:         parsedUUID.String(),
		Name:         name,
	}
	if _, err := db.GetEngine(ctx).Insert(server); err != nil {
		return 0, fmt.Errorf("register Wings server: %w", err)
	}
	return server.ID, nil
}

// CreateServerRecord stores server metadata without verifying against Wings.
// Used for locally defined servers; provisioning on Wings requires the remote API.
func CreateServerRecord(ctx context.Context, nodeID, eggID, allocationID int64, name string) (int64, error) {
	name, err := normalizeServerName(name)
	if err != nil {
		return 0, ErrInvalidServerName
	}
	node, exists, err := GetNodeByID(ctx, nodeID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, fmt.Errorf("Wings node %d does not exist", nodeID)
	}
	_ = node
	server := &WingsServer{
		NodeID:       nodeID,
		EggID:        eggID,
		AllocationID: allocationID,
		UUID:         uuid.New().String(),
		Name:         name,
	}
	if _, err := db.GetEngine(ctx).Insert(server); err != nil {
		return 0, fmt.Errorf("create Wings server record: %w", err)
	}
	return server.ID, nil
}

// RenameServer changes only the server's display name in WebMain.
func RenameServer(ctx context.Context, id int64, name string) error {
	if id <= 0 {
		return errors.New("invalid Wings server ID")
	}
	name, err := normalizeServerName(name)
	if err != nil {
		return err
	}
	server, exists, err := GetServerByID(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Wings server registration %d does not exist", id)
	}
	if server.Name == name {
		return nil
	}
	affected, err := db.GetEngine(ctx).ID(id).Cols("name").Update(&WingsServer{Name: name})
	if err != nil {
		return fmt.Errorf("rename Wings server registration %d: %w", id, err)
	}
	if affected != 1 {
		return fmt.Errorf("rename Wings server registration %d: expected to update one row, updated %d", id, affected)
	}
	return nil
}

func normalizeServerName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 64 {
		return "", ErrInvalidServerName
	}
	return name, nil
}

// ListServers returns registered servers without node credentials.
func ListServers(ctx context.Context) ([]ServerListItem, error) {
	var servers []*WingsServer
	if err := db.GetEngine(ctx).OrderBy("id ASC").Find(&servers); err != nil {
		return nil, fmt.Errorf("list registered Wings servers: %w", err)
	}
	nodes, err := ListNodes(ctx)
	if err != nil {
		return nil, err
	}
	nodeNames := make(map[int64]string, len(nodes))
	for _, node := range nodes {
		nodeNames[node.ID] = node.Name
	}

	serverIDs := make([]int64, 0, len(servers))
	for _, server := range servers {
		serverIDs = append(serverIDs, server.ID)
	}
	accessByServer, err := ListServerAccess(ctx, serverIDs)
	if err != nil {
		return nil, err
	}
	items := make([]ServerListItem, 0, len(servers))
	for _, server := range servers {
		items = append(items, ServerListItem{
			ID:       server.ID,
			NodeID:   server.NodeID,
			NodeName: nodeNames[server.NodeID],
			Name:     server.Name,
			UUID:     server.UUID,
			Accesses: accessByServer[server.ID],
		})
	}
	return items, nil
}

// GetServerByID loads one registered server by its WebMain ID.
func GetServerByID(ctx context.Context, id int64) (*WingsServer, bool, error) {
	if id <= 0 {
		return nil, false, errors.New("invalid Wings server ID")
	}
	server, exists, err := db.GetByID[WingsServer](ctx, id)
	if err != nil {
		return nil, false, fmt.Errorf("get registered Wings server %d: %w", id, err)
	}
	return server, exists, nil
}

// GetServerByUUID returns a registered Wings server by its Wings UUID.
func GetServerByUUID(ctx context.Context, serverUUID string) (*WingsServer, bool, error) {
	serverUUID = strings.TrimSpace(serverUUID)
	if serverUUID == "" {
		return nil, false, fmt.Errorf("invalid Wings server UUID")
	}
	var server WingsServer
	exists, err := db.GetEngine(ctx).Where("uuid = ?", serverUUID).Get(&server)
	if err != nil {
		return nil, false, fmt.Errorf("get Wings server by UUID %s: %w", serverUUID, err)
	}
	if !exists {
		return nil, false, nil
	}
	return &server, true, nil
}

// GetServerClient retrieves the daemon client for a registered server.
func (server *WingsServer) GetServerClient(ctx context.Context) (*wings_service.Client, error) {
	node, exists, err := GetNodeByID(ctx, server.NodeID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("Wings node %d for server %d does not exist", server.NodeID, server.ID)
	}
	return node.GetClient()
}

// HasRemoteCallbackCredentials reports whether Wings can authenticate callbacks.
func (server *WingsServer) HasRemoteCallbackCredentials(ctx context.Context) (bool, error) {
	node, exists, err := GetNodeByID(ctx, server.NodeID)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, fmt.Errorf("Wings node %d for server %d does not exist", server.NodeID, server.ID)
	}
	return node.HasRemoteCallbackCredentials()
}

// DeleteServerRegistration removes the WebMain registration without deleting
// the server data or container from Wings.
func DeleteServerRegistration(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid Wings server ID")
	}
	return db.WithTx(ctx, func(ctx context.Context) error {
		if _, err := db.GetEngine(ctx).Where("server_id = ?", id).Delete(&WingsBackup{}); err != nil {
			return fmt.Errorf("remove Wings backup metadata for registration %d: %w", id, err)
		}
		if _, err := db.GetEngine(ctx).Where("server_id = ?", id).Delete(&WingsServerAccess{}); err != nil {
			return fmt.Errorf("remove Wings server access for registration %d: %w", id, err)
		}
		affected, err := db.GetEngine(ctx).ID(id).Delete(&WingsServer{})
		if err != nil {
			return fmt.Errorf("remove Wings server registration %d: %w", id, err)
		}
		if affected == 0 {
			return fmt.Errorf("Wings server registration %d does not exist", id)
		}
		return nil
	})
}
