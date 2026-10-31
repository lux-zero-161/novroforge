// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_17

import (
	packages_model "novroforge.xyz/models/packages"
	container_module "novroforge.xyz/modules/packages/container"

	"code.novroforge.xyz/xorm/xorm"
	"code.novroforge.xyz/xorm/xorm/schemas"
)

func AddContainerRepositoryProperty(x *xorm.Engine) (err error) {
	if x.Dialect().URI().DBType == schemas.SQLITE {
		_, err = x.Exec("INSERT INTO package_property (ref_type, ref_id, name, value) SELECT ?, p.id, ?, u.lower_name || '/' || p.lower_name FROM package p JOIN `user` u ON p.owner_id = u.id WHERE p.type = ?",
			packages_model.PropertyTypePackage, container_module.PropertyRepository, packages_model.TypeContainer)
	} else {
		_, err = x.Exec("INSERT INTO package_property (ref_type, ref_id, name, value) SELECT ?, p.id, ?, CONCAT(u.lower_name, '/', p.lower_name) FROM package p JOIN `user` u ON p.owner_id = u.id WHERE p.type = ?",
			packages_model.PropertyTypePackage, container_module.PropertyRepository, packages_model.TypeContainer)
	}
	return err
}
