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
	"novroforge.xyz/modules/keying"
	"novroforge.xyz/modules/timeutil"
	wings_service "novroforge.xyz/services/wings"
)

var (
	ErrInvalidNodeName  = errors.New("node name must be between 1 and 64 characters")
	ErrInvalidNodeURL   = errors.New("node URL is invalid")
	ErrInvalidNodeToken = errors.New("node token is invalid")
)

// WingsNode stores connection details for a Wings daemon. The bearer token is
// encrypted at rest and is never included in the node's normal fields.
type WingsNode struct {
	ID             int64
	Name           string             `xorm:"UNIQUE NOT NULL VARCHAR(64)"`
	URL            string             `xorm:"TEXT NOT NULL"`
	EncryptedToken []byte             `xorm:"BLOB"`
	CreatedUnix    timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix    timeutil.TimeStamp `xorm:"updated"`
}

// NodeListItem is the non-sensitive node information shown in admin pages.
type NodeListItem struct {
	ID        int64
	Name      string
	URL       string
	CreatedAt timeutil.TimeStamp
}

func init() {
	db.RegisterModel(new(WingsNode))
}

// CreateNode validates and stores a Wings node, encrypting its token using
// authenticated encryption tied to the database row.
func CreateNode(ctx context.Context, name, nodeURL, token string) error {
	name = strings.TrimSpace(name)
	nodeURL = strings.TrimSpace(nodeURL)
	if name == "" || utf8.RuneCountInString(name) > 64 {
		return ErrInvalidNodeName
	}
	if err := validateNodeToken(token); err != nil {
		return err
	}
	if _, err := wings_service.NewClient(nodeURL, token); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidNodeURL, err)
	}

	return db.WithTx(ctx, func(ctx context.Context) error {
		node := &WingsNode{Name: name, URL: nodeURL}
		if _, err := db.GetEngine(ctx).Insert(node); err != nil {
			return fmt.Errorf("insert Wings node: %w", err)
		}

		node.EncryptedToken = keying.WingsNode.Encrypt([]byte(token), keying.ColumnAndID("encrypted_token", node.ID))
		affected, err := db.GetEngine(ctx).ID(node.ID).Cols("encrypted_token").Update(node)
		if err != nil {
			return fmt.Errorf("encrypt Wings node token: %w", err)
		}
		if affected != 1 {
			return fmt.Errorf("encrypt Wings node token: expected to update one row, updated %d", affected)
		}
		return nil
	})
}

// RotateNodeToken replaces a node's encrypted bearer token without exposing it.
func RotateNodeToken(ctx context.Context, id int64, token string) error {
	if id <= 0 {
		return errors.New("invalid Wings node ID")
	}
	if err := validateNodeToken(token); err != nil {
		return err
	}
	node, exists, err := GetNodeByID(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Wings node %d does not exist", id)
	}
	client, err := wings_service.NewClient(node.URL, token)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidNodeToken, err)
	}
	if _, err := client.GetSystem(ctx); err != nil {
		return fmt.Errorf("verify replacement token for Wings node %d: %w", id, err)
	}
	encryptedToken := keying.WingsNode.Encrypt([]byte(token), keying.ColumnAndID("encrypted_token", node.ID))
	affected, err := db.GetEngine(ctx).ID(node.ID).Cols("encrypted_token").Update(&WingsNode{EncryptedToken: encryptedToken})
	if err != nil {
		return fmt.Errorf("rotate Wings node %d token: %w", id, err)
	}
	if affected != 1 {
		return fmt.Errorf("rotate Wings node %d token: expected to update one row, updated %d", id, affected)
	}
	return nil
}

func validateNodeToken(token string) error {
	if token == "" || len(token) > 4096 || strings.TrimSpace(token) != token {
		return ErrInvalidNodeToken
	}
	for _, char := range token {
		if char < '!' || char > '~' {
			return ErrInvalidNodeToken
		}
	}
	return nil
}

// ListNodes returns node details without encrypted credentials.
func ListNodes(ctx context.Context) ([]NodeListItem, error) {
	var nodes []*WingsNode
	if err := db.GetEngine(ctx).OrderBy("id ASC").Find(&nodes); err != nil {
		return nil, fmt.Errorf("list Wings nodes: %w", err)
	}

	items := make([]NodeListItem, 0, len(nodes))
	for _, node := range nodes {
		items = append(items, NodeListItem{
			ID:        node.ID,
			Name:      node.Name,
			URL:       node.URL,
			CreatedAt: node.CreatedUnix,
		})
	}
	return items, nil
}

// GetNodeByID loads one node configuration without exposing it through the UI.
func GetNodeByID(ctx context.Context, id int64) (*WingsNode, bool, error) {
	if id <= 0 {
		return nil, false, errors.New("invalid Wings node ID")
	}
	node, exists, err := db.GetByID[WingsNode](ctx, id)
	if err != nil {
		return nil, false, fmt.Errorf("get Wings node %d: %w", id, err)
	}
	return node, exists, nil
}

// GetClient decrypts the node credential and returns a configured Wings client.
func (node *WingsNode) GetClient() (*wings_service.Client, error) {
	token, err := keying.WingsNode.Decrypt(node.EncryptedToken, keying.ColumnAndID("encrypted_token", node.ID))
	if err != nil {
		return nil, fmt.Errorf("decrypt Wings node token for node %d: %w", node.ID, err)
	}
	client, err := wings_service.NewClient(node.URL, string(token))
	if err != nil {
		return nil, fmt.Errorf("create Wings client for node %d: %w", node.ID, err)
	}
	return client, nil
}

// DeleteNode removes only the panel's node configuration.
func DeleteNode(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid Wings node ID")
	}
	registeredServers, err := db.GetEngine(ctx).Where("node_id = ?", id).Count(&WingsServer{})
	if err != nil {
		return fmt.Errorf("check Wings node %d server registrations: %w", id, err)
	}
	if registeredServers > 0 {
		return ErrNodeHasServers
	}
	affected, err := db.GetEngine(ctx).ID(id).Delete(&WingsNode{})
	if err != nil {
		return fmt.Errorf("delete Wings node %d: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("Wings node %d does not exist", id)
	}
	return nil
}
