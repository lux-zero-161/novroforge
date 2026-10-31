// SPDX-License-Identifier: MIT

package forgejo_v1_20

import (
	"code.forgejo.org/xorm/xorm"
)

func CreateSemVerTable(x *xorm.Engine) error {
	type NovroForgeSemVer struct {
		Version string
	}

	return x.Sync(new(NovroForgeSemVer))
}
