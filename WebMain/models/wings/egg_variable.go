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

var ErrInvalidEggVariableName = errors.New("egg variable name must be between 1 and 191 characters")

// WingsEggVariable defines an environment variable for an Egg.
type WingsEggVariable struct {
	ID            int64              `xorm:"pk autoincr"`
	EggID         int64              `xorm:"INDEX NOT NULL"`
	Name          string             `xorm:"VARCHAR(191) NOT NULL"`
	Description   string             `xorm:"TEXT"`
	EnvVariable   string             `xorm:"VARCHAR(191) NOT NULL"`
	DefaultValue  string             `xorm:"TEXT"`
	UserViewable  bool               `xorm:"NOT NULL DEFAULT true"`
	UserEditable  bool               `xorm:"NOT NULL DEFAULT true"`
	Rules         string             `xorm:"TEXT"` // comma-separated validation rules
	SortOrder     int                `xorm:"NOT NULL DEFAULT 0"`
	CreatedUnix   timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix   timeutil.TimeStamp `xorm:"updated"`
}

func init() {
	db.RegisterModel(new(WingsEggVariable))
}

func normalizeEggVariableName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 191 {
		return "", ErrInvalidEggVariableName
	}
	return name, nil
}

func CreateEggVariable(ctx context.Context, eggID int64, name, description, envVariable, defaultValue string, userViewable, userEditable bool, rules string) (*WingsEggVariable, error) {
	name, err := normalizeEggVariableName(name)
	if err != nil {
		return nil, err
	}
	envVariable = strings.TrimSpace(envVariable)
	if eggID <= 0 || envVariable == "" {
		return nil, errors.New("invalid Wings egg variable")
	}
	v := &WingsEggVariable{
		EggID:        eggID,
		Name:         name,
		Description:  strings.TrimSpace(description),
		EnvVariable:  envVariable,
		DefaultValue: defaultValue,
		UserViewable: userViewable,
		UserEditable: userEditable,
		Rules:        strings.TrimSpace(rules),
	}
	if _, err := db.GetEngine(ctx).Insert(v); err != nil {
		return nil, fmt.Errorf("create Wings egg variable %q: %w", name, err)
	}
	return v, nil
}

func ListEggVariables(ctx context.Context, eggID int64) ([]*WingsEggVariable, error) {
	vars := make([]*WingsEggVariable, 0)
	q := db.GetEngine(ctx).OrderBy("sort_order ASC, id ASC")
	if eggID > 0 {
		q = q.Where("egg_id = ?", eggID)
	}
	if err := q.Find(&vars); err != nil {
		return nil, fmt.Errorf("list Wings egg variables: %w", err)
	}
	return vars, nil
}

func DeleteEggVariable(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid Wings egg variable ID")
	}
	affected, err := db.GetEngine(ctx).ID(id).Delete(&WingsEggVariable{})
	if err != nil {
		return fmt.Errorf("delete Wings egg variable %d: %w", id, err)
	}
	if affected != 1 {
		return fmt.Errorf("Wings egg variable %d not found", id)
	}
	return nil
}
