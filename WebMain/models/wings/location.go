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

var ErrInvalidLocationName = errors.New("location short name must be between 1 and 191 characters")

// WingsLocation groups nodes by physical location.
type WingsLocation struct {
	ID          int64              `xorm:"pk autoincr"`
	Short       string             `xorm:"VARCHAR(191) NOT NULL"`
	Long        string             `xorm:"TEXT"`
	CreatedUnix timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix timeutil.TimeStamp `xorm:"updated"`
}

func init() {
	db.RegisterModel(new(WingsLocation))
}

func normalizeLocationShort(short string) (string, error) {
	short = strings.TrimSpace(short)
	if short == "" || utf8.RuneCountInString(short) > 191 {
		return "", ErrInvalidLocationName
	}
	return short, nil
}

func CreateLocation(ctx context.Context, short, long string) (*WingsLocation, error) {
	short, err := normalizeLocationShort(short)
	if err != nil {
		return nil, err
	}
	loc := &WingsLocation{Short: short, Long: strings.TrimSpace(long)}
	if _, err := db.GetEngine(ctx).Insert(loc); err != nil {
		return nil, fmt.Errorf("create Wings location %q: %w", short, err)
	}
	return loc, nil
}

func ListLocations(ctx context.Context) ([]*WingsLocation, error) {
	locs := make([]*WingsLocation, 0)
	if err := db.GetEngine(ctx).OrderBy("short ASC").Find(&locs); err != nil {
		return nil, fmt.Errorf("list Wings locations: %w", err)
	}
	return locs, nil
}

func GetLocationByID(ctx context.Context, id int64) (*WingsLocation, bool, error) {
	if id <= 0 {
		return nil, false, errors.New("invalid Wings location ID")
	}
	loc := &WingsLocation{}
	exists, err := db.GetEngine(ctx).ID(id).Get(loc)
	if err != nil {
		return nil, false, fmt.Errorf("get Wings location %d: %w", id, err)
	}
	return loc, exists, nil
}

func DeleteLocation(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid Wings location ID")
	}
	affected, err := db.GetEngine(ctx).ID(id).Delete(&WingsLocation{})
	if err != nil {
		return fmt.Errorf("delete Wings location %d: %w", id, err)
	}
	if affected != 1 {
		return fmt.Errorf("Wings location %d not found", id)
	}
	return nil
}
