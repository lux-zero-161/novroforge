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

var ErrInvalidNestName = errors.New("nest name must be between 1 and 191 characters")

// WingsNest groups related Eggs (server types).
type WingsNest struct {
	ID          int64              `xorm:"pk autoincr"`
	Name        string             `xorm:"VARCHAR(191) NOT NULL"`
	Description string             `xorm:"TEXT"`
	CreatedUnix timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
}

func init() {
	db.RegisterModel(new(WingsNest))
}

func normalizeNestName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 191 {
		return "", ErrInvalidNestName
	}
	return name, nil
}

func CreateNest(ctx context.Context, name, description string) (*WingsNest, error) {
	name, err := normalizeNestName(name)
	if err != nil {
		return nil, err
	}
	nest := &WingsNest{Name: name, Description: strings.TrimSpace(description)}
	if _, err := db.GetEngine(ctx).Insert(nest); err != nil {
		return nil, fmt.Errorf("create Wings nest %q: %w", name, err)
	}
	return nest, nil
}

func ListNests(ctx context.Context) ([]*WingsNest, error) {
	nests := make([]*WingsNest, 0)
	if err := db.GetEngine(ctx).OrderBy("name ASC").Find(&nests); err != nil {
		return nil, fmt.Errorf("list Wings nests: %w", err)
	}
	return nests, nil
}

func GetNestByID(ctx context.Context, id int64) (*WingsNest, bool, error) {
	if id <= 0 {
		return nil, false, errors.New("invalid Wings nest ID")
	}
	nest := &WingsNest{}
	exists, err := db.GetEngine(ctx).ID(id).Get(nest)
	if err != nil {
		return nil, false, fmt.Errorf("get Wings nest %d: %w", id, err)
	}
	return nest, exists, nil
}

func DeleteNest(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid Wings nest ID")
	}
	affected, err := db.GetEngine(ctx).ID(id).Delete(&WingsNest{})
	if err != nil {
		return fmt.Errorf("delete Wings nest %d: %w", id, err)
	}
	if affected != 1 {
		return fmt.Errorf("Wings nest %d not found", id)
	}
	return nil
}
