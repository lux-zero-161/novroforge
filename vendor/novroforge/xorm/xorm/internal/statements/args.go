// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package statements

import (
	"xorm.io/builder"
)

// WriteArg writes an arg
func (statement *Statement) WriteArg(w *builder.BytesWriter, arg any) error {
	switch argv := arg.(type) {
	case *builder.Builder:
		if _, err := w.WriteString("("); err != nil {
			return err
		}
		if err := argv.WriteTo(w); err != nil {
			return err
		}
		if _, err := w.WriteString(")"); err != nil {
			return err
		}
	default:
		if err := w.WriteByte('?'); err != nil {
			return err
		}
		w.Append(arg)
	}
	return nil
}

// WriteArgs writes args
func (statement *Statement) WriteArgs(w *builder.BytesWriter, args []any) error {
	for i, arg := range args {
		if err := statement.WriteArg(w, arg); err != nil {
			return err
		}

		if i+1 != len(args) {
			if _, err := w.WriteString(","); err != nil {
				return err
			}
		}
	}
	return nil
}
