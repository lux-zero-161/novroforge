// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package wings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"novroforge.xyz/models/db"
	"novroforge.xyz/modules/timeutil"
)

var ErrInvalidEggName = errors.New("egg name must be between 1 and 191 characters")

// WingsEgg describes a server type (template) that can be installed on Wings.
type WingsEgg struct {
	ID           int64              `xorm:"pk autoincr"`
	NestID       int64              `xorm:"INDEX NOT NULL"`
	Name         string             `xorm:"VARCHAR(191) NOT NULL"`
	Description  string             `xorm:"TEXT"`
	DockerImage  string             `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
	Startup      string             `xorm:"TEXT"`
	Config       string             `xorm:"TEXT"` // JSON: files, startup, stop, logs, etc.
	CreatedUnix  timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix  timeutil.TimeStamp `xorm:"updated"`
}

// EggConfig holds the structured configuration parsed from WingsEgg.Config.
type EggConfig struct {
	Files   map[string]any `json:"files"`
	Startup map[string]any `json:"startup"`
	Stop    string         `json:"stop"`
	Logs    []string       `json:"logs"`
}

func init() {
	db.RegisterModel(new(WingsEgg))
}

func normalizeEggName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 191 {
		return "", ErrInvalidEggName
	}
	return name, nil
}

func CreateEgg(ctx context.Context, nestID int64, name, description, dockerImage, startup string, config EggConfig) (*WingsEgg, error) {
	name, err := normalizeEggName(name)
	if err != nil {
		return nil, err
	}
	if nestID <= 0 {
		return nil, errors.New("invalid Wings nest ID")
	}
	cfgBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("encode egg config: %w", err)
	}
	egg := &WingsEgg{
		NestID:      nestID,
		Name:        name,
		Description: strings.TrimSpace(description),
		DockerImage: strings.TrimSpace(dockerImage),
		Startup:     strings.TrimSpace(startup),
		Config:      string(cfgBytes),
	}
	if _, err := db.GetEngine(ctx).Insert(egg); err != nil {
		return nil, fmt.Errorf("create Wings egg %q: %w", name, err)
	}
	return egg, nil
}

func ListEggs(ctx context.Context, nestID int64) ([]*WingsEgg, error) {
	eggs := make([]*WingsEgg, 0)
	q := db.GetEngine(ctx).OrderBy("name ASC")
	if nestID > 0 {
		q = q.Where("nest_id = ?", nestID)
	}
	if err := q.Find(&eggs); err != nil {
		return nil, fmt.Errorf("list Wings eggs: %w", err)
	}
	return eggs, nil
}

func GetEggByID(ctx context.Context, id int64) (*WingsEgg, bool, error) {
	if id <= 0 {
		return nil, false, errors.New("invalid Wings egg ID")
	}
	egg := &WingsEgg{}
	exists, err := db.GetEngine(ctx).ID(id).Get(egg)
	if err != nil {
		return nil, false, fmt.Errorf("get Wings egg %d: %w", id, err)
	}
	return egg, exists, nil
}

func DeleteEgg(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid Wings egg ID")
	}
	affected, err := db.GetEngine(ctx).ID(id).Delete(&WingsEgg{})
	if err != nil {
		return fmt.Errorf("delete Wings egg %d: %w", id, err)
	}
	if affected != 1 {
		return fmt.Errorf("Wings egg %d not found", id)
	}
	return nil
}

// ParseConfig decodes the stored JSON configuration.
func (egg *WingsEgg) ParseConfig() (EggConfig, error) {
	var cfg EggConfig
	if strings.TrimSpace(egg.Config) == "" {
		return cfg, nil
	}
	if err := json.Unmarshal([]byte(egg.Config), &cfg); err != nil {
		return cfg, fmt.Errorf("parse Wings egg config: %w", err)
	}
	return cfg, nil
}
