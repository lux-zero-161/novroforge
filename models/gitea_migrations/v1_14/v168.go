// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_14

import "novroforge/xorm/xorm"

func RecreateUserTableToFixDefaultValues(_ *xorm.Engine) error {
	return nil
}
