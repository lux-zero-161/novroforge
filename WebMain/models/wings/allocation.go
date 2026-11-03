// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package wings

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"novroforge.xyz/models/db"
	"novroforge.xyz/modules/timeutil"
)

var ErrInvalidAllocationIP = errors.New("allocation IP must be a valid IP address")

// WingsAllocation represents a network port allocation on a node.
type WingsAllocation struct {
	ID          int64              `xorm:"pk autoincr"`
	NodeID      int64              `xorm:"INDEX NOT NULL"`
	IP          string             `xorm:"VARCHAR(45) NOT NULL"`
	Port        int                `xorm:"NOT NULL"`
	IPAlias     string             `xorm:"VARCHAR(255)"`
	Assigned    bool               `xorm:"NOT NULL DEFAULT false"`
	ServerID    int64              `xorm:"INDEX"`
	CreatedUnix timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
}

func init() {
	db.RegisterModel(new(WingsAllocation))
}

func validateAllocationIP(ip string) error {
	if net.ParseIP(strings.TrimSpace(ip)) == nil {
		return ErrInvalidAllocationIP
	}
	return nil
}

func CreateAllocation(ctx context.Context, nodeID int64, ip string, port int, alias string) (*WingsAllocation, error) {
	ip = strings.TrimSpace(ip)
	if nodeID <= 0 || port <= 0 || port > 65535 {
		return nil, errors.New("invalid Wings allocation")
	}
	if err := validateAllocationIP(ip); err != nil {
		return nil, err
	}
	a := &WingsAllocation{NodeID: nodeID, IP: ip, Port: port, IPAlias: strings.TrimSpace(alias)}
	if _, err := db.GetEngine(ctx).Insert(a); err != nil {
		return nil, fmt.Errorf("create Wings allocation: %w", err)
	}
	return a, nil
}

func ListAllocations(ctx context.Context, nodeID int64) ([]*WingsAllocation, error) {
	allocs := make([]*WingsAllocation, 0)
	q := db.GetEngine(ctx).OrderBy("ip ASC, port ASC")
	if nodeID > 0 {
		q = q.Where("node_id = ?", nodeID)
	}
	if err := q.Find(&allocs); err != nil {
		return nil, fmt.Errorf("list Wings allocations: %w", err)
	}
	return allocs, nil
}

func GetAllocationByID(ctx context.Context, id int64) (*WingsAllocation, bool, error) {
	if id <= 0 {
		return nil, false, errors.New("invalid Wings allocation ID")
	}
	a := &WingsAllocation{}
	exists, err := db.GetEngine(ctx).ID(id).Get(a)
	if err != nil {
		return nil, false, fmt.Errorf("get Wings allocation %d: %w", id, err)
	}
	return a, exists, nil
}

func DeleteAllocation(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid Wings allocation ID")
	}
	affected, err := db.GetEngine(ctx).ID(id).Where("assigned = ?", false).Delete(&WingsAllocation{})
	if err != nil {
		return fmt.Errorf("delete Wings allocation %d: %w", id, err)
	}
	if affected != 1 {
		return errors.New("Wings allocation not found or currently assigned")
	}
	return nil
}
