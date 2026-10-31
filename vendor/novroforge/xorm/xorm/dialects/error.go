// Copyright 2025 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dialects

import (
	"errors"
)

var (
	// ErrForeignKeyViolation occurs when a modification is attempted that violates the DB's foreign keys
	ErrForeignKeyViolation = errors.New("foreign key constraint violation")
	// ErrUniqueConstraintViolation occurs when a modification attempts to add a duplicate value to a table that
	// violates a unique constraint, including a primary key
	ErrUniqueConstraintViolation = errors.New("unique constraint violation")
	// ErrDeadlock occurs when a deadlock in the database occurs.
	ErrDeadlock = errors.New("database deadlock occurred")
)
