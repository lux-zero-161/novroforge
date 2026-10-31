// Copyright 2023 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package statements

import (
	"fmt"

	"xorm.io/builder"
)

func (statement *Statement) writePagination(bw *builder.BytesWriter) error {
	return statement.writeLimitOffset(bw)
}

func (statement *Statement) writeLimitOffset(w builder.Writer) error {
	if statement.Start > 0 {
		if statement.LimitN != nil {
			_, err := fmt.Fprintf(w, " LIMIT %v OFFSET %v", *statement.LimitN, statement.Start)
			return err
		}
		_, err := fmt.Fprintf(w, " OFFSET %v", statement.Start)
		return err
	}
	if statement.LimitN != nil {
		_, err := fmt.Fprint(w, " LIMIT ", *statement.LimitN)
		return err
	}
	// no limit statement
	return nil
}
