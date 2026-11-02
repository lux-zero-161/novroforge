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
func CreateServer(ctx context.Context, nodeID int64, name, serverUUID string) error {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 64 {
		return ErrInvalidServerName
	}
	parsedUUID, err := uuid.Parse(strings.TrimSpace(serverUUID))
	if err != nil {
		return ErrInvalidServerUUID
	}
	node, exists, err := GetNodeByID(ctx, nodeID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Wings node %d does not exist", nodeID)
	}
	client, err := node.GetClient()
	if err != nil {
		return err
	}
	if _, err := client.GetServer(ctx, parsedUUID.String()); err != nil {
		return fmt.Errorf("verify server on Wings node: %w", err)
	}

	server := &WingsServer{
		NodeID: nodeID,
		UUID:   parsedUUID.String(),
		Name:   name,
	}
	if _, err := db.GetEngine(ctx).Insert(server); err != nil {
		return fmt.Errorf("register Wings server: %w", err)
	}
	return nil
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

// DeleteServerRegistration removes the WebMain registration without deleting
// the server data or container from Wings.
func DeleteServerRegistration(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid Wings server ID")
	}
	return db.WithTx(ctx, func(ctx context.Context) error {
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
