// Copyright 2024 The NovroForge Authors c/o Codeberg e.V.. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_22

import (
	"code.novroforge.xyz/xorm/xorm"
)

type RepoFlag struct {
	ID     int64  `xorm:"pk autoincr"`
	RepoID int64  `xorm:"UNIQUE(s) INDEX"`
	Name   string `xorm:"UNIQUE(s) INDEX"`
}

func (RepoFlag) TableName() string {
	return "novroforge_repo_flag"
}

func CreateRepoFlagTable(x *xorm.Engine) error {
	return x.Sync(new(RepoFlag))
}
