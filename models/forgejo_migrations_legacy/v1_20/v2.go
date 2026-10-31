// SPDX-License-Identifier: MIT

package novroforge_v1_20

import (
	"code.novroforge.xyz/xorm/xorm"
)

func CreateSemVerTable(x *xorm.Engine) error {
	type NovroForgeSemVer struct {
		Version string
	}

	return x.Sync(new(NovroForgeSemVer))
}
