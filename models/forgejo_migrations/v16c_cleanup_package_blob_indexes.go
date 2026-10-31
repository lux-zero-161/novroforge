// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package novroforge_migrations

import (
	"code.novroforge.xyz/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "cleanup extra indexes on package_blob",
		Upgrade:     cleanupPackageBlobIndexes,
	})
}

func cleanupPackageBlobIndexes(x *xorm.Engine) error {
	for _, idx := range []string{
		"IDX_package_blob_hash_blake2b",
		"IDX_package_blob_hash_md5",
		"IDX_package_blob_hash_sha1",
		"IDX_package_blob_hash_sha256",
		"IDX_package_blob_hash_sha512",
	} {
		err := dropIndexIfExists(x, "package_blob", idx)
		if err != nil {
			return err
		}
	}
	return nil
}
