// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_16

import (
	"code.novroforge.xyz/xorm/xorm"
)

func UseBase32HexForCredIDInWebAuthnCredential(x *xorm.Engine) error {
	// noop
	return nil
}
