// Copyright 2016 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"database/sql"
	"strings"

	"code.forgejo.org/xorm/xorm/core"
)

func (session *Session) queryPreprocess(sqlStr *string, paramStr ...any) {
	for _, filter := range session.engine.dialect.Filters() {
		*sqlStr = filter.Do(session.ctx, *sqlStr)
	}

	session.lastSQL = *sqlStr
	session.lastSQLArgs = paramStr
}

func (session *Session) errorPostprocess(err error) error {
	if err != nil {
		err = session.engine.driver.StandardizeError(err)
	}
	return err
}

func (session *Session) queryRows(sqlStr string, args ...any) (*core.Rows, error) {
	defer session.resetStatement()
	if session.statement.LastError != nil {
		return nil, session.statement.LastError
	}

	session.queryPreprocess(&sqlStr, args...)

	session.lastSQL = sqlStr
	session.lastSQLArgs = args

	if session.isAutoCommit {
		var db *core.DB
		if session.sessionType == groupSession && strings.EqualFold(strings.TrimSpace(sqlStr)[:6], "select") && !session.statement.IsForUpdate {
			db = session.engine.engineGroup.Slave().DB()
		} else {
			db = session.DB()
		}

		if session.prepareStmt {
			// don't clear stmt since session will cache them
			stmt, err := session.doPrepare(db, sqlStr)
			if err != nil {
				return nil, err
			}

			result, err := stmt.QueryContext(session.ctx, args...)
			return result, session.errorPostprocess(err)
		}

		result, err := db.QueryContext(session.ctx, sqlStr, args...)
		return result, session.errorPostprocess(err)
	}

	if session.prepareStmt {
		stmt, err := session.doPrepareTx(sqlStr)
		if err != nil {
			return nil, err
		}

		result, err := stmt.QueryContext(session.ctx, args...)
		return result, session.errorPostprocess(err)
	}

	result, err := session.tx.QueryContext(session.ctx, sqlStr, args...)
	return result, session.errorPostprocess(err)
}

func (session *Session) queryRow(sqlStr string, args ...any) *core.Row {
	return core.NewRow(session.queryRows(sqlStr, args...))
}

// Query runs a raw sql and return records as []map[string][]byte
func (session *Session) Query(sqlOrArgs ...any) ([]map[string][]byte, error) {
	if session.isAutoClose {
		defer session.Close()
	}

	sqlStr, args, err := session.statement.GenQuerySQL(sqlOrArgs...)
	if err != nil {
		return nil, err
	}

	rows, err := session.queryRows(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return session.engine.scanByteMaps(rows)
}

// QueryString runs a raw sql and return records as []map[string]string
func (session *Session) QueryString(sqlOrArgs ...any) ([]map[string]string, error) {
	if session.isAutoClose {
		defer session.Close()
	}

	sqlStr, args, err := session.statement.GenQuerySQL(sqlOrArgs...)
	if err != nil {
		return nil, err
	}

	rows, err := session.queryRows(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return session.engine.ScanStringMaps(rows)
}

// QuerySliceString runs a raw sql and return records as [][]string
func (session *Session) QuerySliceString(sqlOrArgs ...any) ([][]string, error) {
	if session.isAutoClose {
		defer session.Close()
	}

	sqlStr, args, err := session.statement.GenQuerySQL(sqlOrArgs...)
	if err != nil {
		return nil, err
	}

	rows, err := session.queryRows(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return session.engine.ScanStringSlices(rows)
}

// QueryInterface runs a raw sql and return records as []map[string]interface{}
func (session *Session) QueryInterface(sqlOrArgs ...any) ([]map[string]any, error) {
	if session.isAutoClose {
		defer session.Close()
	}

	sqlStr, args, err := session.statement.GenQuerySQL(sqlOrArgs...)
	if err != nil {
		return nil, err
	}

	rows, err := session.queryRows(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return session.engine.ScanInterfaceMaps(rows)
}

func (session *Session) exec(sqlStr string, args ...any) (sql.Result, error) {
	defer session.resetStatement()

	session.queryPreprocess(&sqlStr, args...)

	session.lastSQL = sqlStr
	session.lastSQLArgs = args

	if !session.isAutoCommit {
		if session.prepareStmt {
			stmt, err := session.doPrepareTx(sqlStr)
			if err != nil {
				return nil, err
			}
			return stmt.ExecContext(session.ctx, args...)
		}
		result, err := session.tx.ExecContext(session.ctx, sqlStr, args...)
		return result, session.errorPostprocess(err)
	}

	if session.prepareStmt {
		stmt, err := session.doPrepare(session.DB(), sqlStr)
		if err != nil {
			return nil, err
		}
		result, err := stmt.ExecContext(session.ctx, args...)
		return result, session.errorPostprocess(err)
	}

	result, err := session.DB().ExecContext(session.ctx, sqlStr, args...)
	return result, session.errorPostprocess(err)
}

// Exec raw sql
func (session *Session) Exec(sqlOrArgs ...any) (sql.Result, error) {
	if session.isAutoClose {
		defer session.Close()
	}

	if len(sqlOrArgs) == 0 {
		return nil, ErrUnSupportedType
	}

	sqlStr, args, err := session.statement.ConvertSQLOrArgs(sqlOrArgs...)
	if err != nil {
		return nil, err
	}

	return session.exec(sqlStr, args...)
}
