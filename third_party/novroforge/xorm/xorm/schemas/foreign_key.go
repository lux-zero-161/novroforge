// Copyright 2025 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schemas

import "fmt"

// ForeignKey represents a table's foreign key to another table.  Only a single-column FK reference is supported.  xorm
// only supports foreign keys that are...
//
// - ON DELETE RESTRICT
// - ON UPDATE RESTRICT
// - NOT DEFERRABLE
// - INITIALLY IMMEDIATE
//
// ... and so none of these options are represented due to the lack of support.
type ForeignKey struct {
	Name            string // name of the constraint in the DB
	SourceFieldName string
	TargetTableName string
	TargetFieldName string
}

func AutoNamedForeignKey(sourceTableName, sourceFieldName, targetTableName, targetFieldName string) ForeignKey {
	return ForeignKey{
		Name:            fmt.Sprintf("%s_%s_fkey", sourceTableName, sourceFieldName),
		SourceFieldName: sourceFieldName,
		TargetTableName: targetTableName,
		TargetFieldName: targetFieldName,
	}
}

// Equal return true if the two FKs are equal
func (fk *ForeignKey) Equal(dst *ForeignKey) bool {
	// Simple implementation since all fields are comparable right now
	return fk == dst
}
