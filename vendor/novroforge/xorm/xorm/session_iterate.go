// Copyright 2016 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"reflect"
)

// IterFunc only use by Iterate
type IterFunc func(idx int, bean any) error

// Rows return sql.Rows compatible Rows obj, as a forward Iterator object for iterating record by record, bean's non-empty fields
// are conditions.
func (session *Session) Rows(bean any) (*Rows, error) {
	return newRows(session, bean)
}

// Iterate record by record handle records from table, condiBeans's non-empty fields
// are conditions. beans could be []Struct, []*Struct, map[int64]Struct
// map[int64]*Struct
func (session *Session) Iterate(bean any, fun IterFunc) error {
	if session.isAutoClose {
		defer session.Close()
	}

	session.autoResetStatement = false
	defer func() {
		session.autoResetStatement = true
		session.resetStatement()
	}()

	if session.statement.LastError != nil {
		return session.statement.LastError
	}

	rows, err := session.Rows(bean)
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		b := reflect.New(rows.beanType).Interface()
		err = rows.Scan(b)
		if err != nil {
			return err
		}
		err = fun(i, b)
		if err != nil {
			return err
		}
		i++
	}
	return rows.Err()
}
