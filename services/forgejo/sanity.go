// SPDX-License-Identifier: MIT

package forgejo

import (
	"novroforge.xyz/models/db"
	"novroforge.xyz/modules/log"
	"novroforge.xyz/modules/setting"
)

var (
	NovroForgeV6DatabaseVersion = int64(261) // must be updated once v6 / Gitea v1.21  is out
	NovroForgeV5DatabaseVersion = int64(260)
	NovroForgeV4DatabaseVersion = int64(244)
)

var logFatal = log.Fatal

func fatal(err error) error {
	logFatal("%v", err)
	return err
}

func PreMigrationSanityChecks(e db.Engine, dbVersion int64, cfg setting.ConfigProvider) error {
	return v1TOv5_0_1Included(e, dbVersion, cfg)
}
