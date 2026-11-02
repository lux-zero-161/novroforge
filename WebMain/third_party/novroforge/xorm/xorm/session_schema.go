// Copyright 2016 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"

	"novroforge/xorm/xorm/core"
	"novroforge/xorm/xorm/dialects"
	"novroforge/xorm/xorm/internal/utils"
	"novroforge/xorm/xorm/schemas"
)

// Ping test if database is ok
func (session *Session) Ping() error {
	if session.isAutoClose {
		defer session.Close()
	}

	session.engine.logger.Infof("PING DATABASE %v", session.engine.DriverName())
	return session.DB().PingContext(session.ctx)
}

// CreateTable create a table according a bean
func (session *Session) CreateTable(bean any) error {
	if session.isAutoClose {
		defer session.Close()
	}

	return session.createTable(bean)
}

func (session *Session) createTable(bean any) error {
	if err := session.statement.SetRefBean(bean); err != nil {
		return err
	}

	session.statement.RefTable.StoreEngine = session.statement.StoreEngine
	session.statement.RefTable.Charset = session.statement.Charset
	tableName := session.statement.TableName()
	refTable := session.statement.RefTable
	if refTable.AutoIncrement != "" && session.engine.dialect.Features().AutoincrMode == dialects.SequenceAutoincrMode {
		sqlStr, err := session.engine.dialect.CreateSequenceSQL(context.Background(), session.engine.db, utils.SeqName(tableName))
		if err != nil {
			return err
		}
		if _, err := session.exec(sqlStr); err != nil {
			return err
		}
	}

	sqlStr, _, err := session.engine.dialect.CreateTableSQL(refTable, tableName)
	if err != nil {
		return err
	}
	if _, err := session.exec(sqlStr); err != nil {
		return err
	}

	return nil
}

// CreateIndexes create indexes
func (session *Session) CreateIndexes(bean any) error {
	if session.isAutoClose {
		defer session.Close()
	}

	return session.createIndexes(bean)
}

func (session *Session) createIndexes(bean any) error {
	if err := session.statement.SetRefBean(bean); err != nil {
		return err
	}

	sqls := session.statement.GenIndexSQL()
	for _, sqlStr := range sqls {
		_, err := session.exec(sqlStr)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateUniques create uniques
func (session *Session) CreateUniques(bean any) error {
	if session.isAutoClose {
		defer session.Close()
	}
	return session.createUniques(bean)
}

func (session *Session) createUniques(bean any) error {
	if err := session.statement.SetRefBean(bean); err != nil {
		return err
	}

	sqls := session.statement.GenUniqueSQL()
	for _, sqlStr := range sqls {
		_, err := session.exec(sqlStr)
		if err != nil {
			return err
		}
	}
	return nil
}

// DropIndexes drop indexes
func (session *Session) DropIndexes(bean any) error {
	if session.isAutoClose {
		defer session.Close()
	}

	return session.dropIndexes(bean)
}

func (session *Session) dropIndexes(bean any) error {
	if err := session.statement.SetRefBean(bean); err != nil {
		return err
	}

	sqls := session.statement.GenDelIndexSQL()
	for _, sqlStr := range sqls {
		_, err := session.exec(sqlStr)
		if err != nil {
			return err
		}
	}
	return nil
}

// DropTable drop table will drop table if exist, if drop failed, it will return error
func (session *Session) DropTable(beanOrTableName any) error {
	if session.isAutoClose {
		defer session.Close()
	}

	return session.dropTable(beanOrTableName)
}

func (session *Session) dropTable(beanOrTableName any) error {
	tableName := session.engine.TableName(beanOrTableName)
	sqlStr, checkIfExist := session.engine.dialect.DropTableSQL(session.engine.TableName(tableName, true))
	if !checkIfExist {
		exist, err := session.engine.dialect.IsTableExist(session.ctx, session.getQueryer(), tableName)
		if err != nil {
			return err
		}
		checkIfExist = exist
	}

	if !checkIfExist {
		return nil
	}
	if _, err := session.exec(sqlStr); err != nil {
		return err
	}

	if session.engine.dialect.Features().AutoincrMode == dialects.IncrAutoincrMode {
		return nil
	}

	seqName := utils.SeqName(tableName)
	exist, err := session.engine.dialect.IsSequenceExist(session.ctx, session.getQueryer(), seqName)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}

	sqlStr, err = session.engine.dialect.DropSequenceSQL(seqName)
	if err != nil {
		return err
	}
	_, err = session.exec(sqlStr)
	return err
}

// IsTableExist if a table is exist
func (session *Session) IsTableExist(beanOrTableName any) (bool, error) {
	if session.isAutoClose {
		defer session.Close()
	}

	tableName := session.engine.TableName(beanOrTableName)

	return session.isTableExist(tableName)
}

func (session *Session) isTableExist(tableName string) (bool, error) {
	return session.engine.dialect.IsTableExist(session.ctx, session.getQueryer(), tableName)
}

// IsTableEmpty if table have any records
func (session *Session) IsTableEmpty(bean any) (bool, error) {
	if session.isAutoClose {
		defer session.Close()
	}
	return session.isTableEmpty(session.engine.TableName(bean))
}

func (session *Session) isTableEmpty(tableName string) (bool, error) {
	var total int64
	sqlStr := fmt.Sprintf("select count(*) from %s", session.engine.Quote(session.engine.TableName(tableName, true)))
	err := session.queryRow(sqlStr).Scan(&total)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return true, err
	}

	return total == 0, nil
}

func (session *Session) addColumn(colName string) error {
	col := session.statement.RefTable.GetColumn(colName)
	sql := session.engine.dialect.AddColumnSQL(session.statement.TableName(), col)
	_, err := session.exec(sql)
	return err
}

func (session *Session) addIndex(tableName, idxName string) error {
	index := session.statement.RefTable.Indexes[idxName]
	sqlStr := session.engine.dialect.CreateIndexSQL(tableName, index)
	_, err := session.exec(sqlStr)
	return err
}

func (session *Session) addUnique(tableName, uqeName string) error {
	index := session.statement.RefTable.Indexes[uqeName]
	sqlStr := session.engine.dialect.CreateIndexSQL(tableName, index)
	_, err := session.exec(sqlStr)
	return err
}

// ImportFile SQL DDL file
func (session *Session) ImportFile(ddlPath string) ([]sql.Result, error) {
	file, err := os.Open(ddlPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return session.Import(file)
}

// Import SQL DDL from io.Reader
func (session *Session) Import(r io.Reader) ([]sql.Result, error) {
	var (
		results       []sql.Result
		lastError     error
		inSingleQuote bool
		startComment  bool
	)

	scanner := bufio.NewScanner(r)
	semiColSpliter := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		oriInSingleQuote := inSingleQuote
		for i, b := range data {
			if startComment {
				if b == '\n' {
					startComment = false
				}
			} else {
				if !inSingleQuote && i > 0 && data[i-1] == '-' && data[i] == '-' {
					startComment = true
					continue
				}

				if b == '\'' {
					inSingleQuote = !inSingleQuote
				}
				if !inSingleQuote && b == ';' {
					return i + 1, data[0:i], nil
				}
			}
		}
		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), data, nil
		}
		inSingleQuote = oriInSingleQuote
		// Request more data.
		return 0, nil, nil
	}

	scanner.Split(semiColSpliter)

	for scanner.Scan() {
		query := strings.Trim(scanner.Text(), " \t\n\r")
		if len(query) > 0 {
			result, err := session.Exec(query)
			if err != nil {
				return nil, err
			}
			results = append(results, result)
		}
	}

	return results, lastError
}

func (session *Session) IndexHint(op, forType, indexerOrColName string) *Session {
	session.statement.IndexHint(op, forType, indexerOrColName)
	return session
}

// AlterTable applies table schema modifications.  Although `Engine.Sync()` is typically used for xorm schema
// modifications, `AlterTable` is intended to expose some modification capabilities that an xorm client might need a
// little more fine-grained control over.  Currently only FK modifications are supported, but a long-term vision would
// be to make all the Sync() operations use `AlterTable(...)` under the hood, exposing providing both automatic
// bean-generated syncs, and complete access to finer grained modifications.
//
// You must provide the full bean definition to `AlterTable`!  xorm's for SQLite clone the table, and will lose any
// schema definition facets that xorm can't parse from reflection and the xorm field tags.  Other schema definitions
// that xorm doesn't understand, for example, a CHECK constraint, a concept not implemented/supported by xorm, would be
// lost during the AlterTable operation.
//
// xorm's implementation of `AlterTable` for SQLite also requires that any views that refer to the table will need to be
// recreated.
func (session *Session) AlterTable(bean any, modifications ...schemas.TableModification) error {
	if len(modifications) == 0 {
		return nil
	}

	reflectedTable, err := session.engine.TableInfo(bean)
	if err != nil {
		return err
	}

	return session.internalAlterTable(bean, reflectedTable, modifications...)
}

func (session *Session) internalAlterTable(bean any, originalTable *schemas.Table, modifications ...schemas.TableModification) error {
	if len(modifications) == 0 {
		return nil
	}

	var tbName string
	if len(session.statement.AltTableName) > 0 {
		tbName = session.statement.AltTableName
	} else {
		tbName = session.engine.TableName(bean)
	}
	tbNameWithSchema := session.engine.tbNameWithSchema(tbName)

	script, err := session.engine.dialect.GetAlterTableSQL(originalTable, tbNameWithSchema, modifications...)
	if err != nil {
		return err
	}

	hasTxStart := false
	hasTxCommit := false
	var deferredOps []dialects.AlterTableOp
	for _, op := range script {
		switch op.OpType {
		case dialects.TxStartOp:
			hasTxStart = true
		case dialects.TxCommitOp:
			hasTxCommit = true
		case dialects.DeferredSQLOp:
			deferredOps = append(deferredOps, op)
		}
	}
	defer func() {
		for _, op := range deferredOps {
			if _, err := session.exec(op.SQL); err != nil {
				session.engine.logger.Errorf("alter table deferred op %q failed: %v", op.SQL, err)
			}
		}
	}()

	// Create a new transaction for the alter unless we're in a transaction from the caller.  For most dialects the
	// alter transaction will be a single statement, but for SQLite it will be more complex (and the documention says to
	// use a TX for their procedure).
	manageTx := !session.IsInTx()
	if manageTx && !hasTxStart {
		err = session.Begin()
		if err != nil {
			return err
		}
	}

	for _, op := range script {
		switch op.OpType {
		case dialects.TxStartOp:
			if manageTx {
				err = session.Begin()
			}
		case dialects.TxCommitOp:
			if manageTx {
				err = session.Commit()
			}
		case dialects.SQLOp:
			if op.QueryCB != nil {
				var rows *core.Rows
				rows, err = session.queryRows(op.SQL)
				if err == nil {
					err = op.QueryCB(rows.Rows)
				}
			} else {
				_, err = session.exec(op.SQL)
			}
		case dialects.DeferredSQLOp:
			// executed in deferred function
		default:
			err = fmt.Errorf("unknown alter table operation: %v", op.OpType)
		}
		if err != nil {
			if manageTx {
				_ = session.Rollback()
			}
			return err
		}
	}
	if manageTx && !hasTxCommit {
		err = session.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}
