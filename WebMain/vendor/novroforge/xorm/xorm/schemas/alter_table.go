// Copyright 2025 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schemas

type TableModification interface {
	// Seal the interface with a private method
	isTableModification()
}

type AddForeignKey struct {
	ForeignKey ForeignKey
}

func (a AddForeignKey) isTableModification() {}

type DropForeignKey struct {
	ForeignKey ForeignKey
}

func (a DropForeignKey) isTableModification() {}
