// Copyright 2019 Lunny Xiao. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package levelqueue

import "errors"

var (
	// ErrNotFound means no elements in queue
	ErrNotFound = errors.New("no key found")

	ErrAlreadyInQueue = errors.New("value already in queue")

	// ErrLowIDNotExists means the value of the "low" position does not point to
	// a item that's stored in LevelDB. The next five values also have been
	// checked and did not contain a item.
	ErrLowIDNotExists = errors.New("the value of 'low' does not point to a item that's stored in LevelDB")
)
