// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package asymkey

import (
	"context"

	"novroforge.xyz/models"
	asymkey_model "novroforge.xyz/models/asymkey"
	"novroforge.xyz/models/db"
)

// DeleteDeployKey deletes deploy key from its repository authorized_keys file if needed.
func DeleteDeployKey(ctx context.Context, id, repoID int64) error {
	if err := db.WithTx(ctx, func(ctx context.Context) error {
		return models.DeleteDeployKey(ctx, id, repoID)
	}); err != nil {
		return err
	}

	return asymkey_model.RewriteAllPublicKeys(ctx)
}
