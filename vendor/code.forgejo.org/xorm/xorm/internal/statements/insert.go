// Copyright 2020 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package statements

import (
	"errors"
	"fmt"
	"strings"

	"code.forgejo.org/xorm/xorm/schemas"
	"xorm.io/builder"
)

func (statement *Statement) writeInsertOutput(buf *strings.Builder, table *schemas.Table) error {
	return nil
}

// GenInsertSQL generates insert beans SQL
func (statement *Statement) GenInsertSQL(colNames []string, args []any) (string, []any, error) {
	var (
		buf       = builder.NewWriter()
		exprs     = statement.ExprColumns
		table     = statement.RefTable
		tableName = statement.TableName()
	)

	if _, err := buf.WriteString("INSERT INTO "); err != nil {
		return "", nil, err
	}

	statement.dialect.Quoter().QuoteTo(buf.Builder, tableName)

	hasInsertColumns := len(colNames) > 0

	if !hasInsertColumns {
		if statement.dialect.URI().DBType == schemas.MYSQL {
			if _, err := buf.WriteString(" VALUES ()"); err != nil {
				return "", nil, err
			}
		} else {
			if err := statement.writeInsertOutput(buf.Builder, table); err != nil {
				return "", nil, err
			}
			if _, err := buf.WriteString(" DEFAULT VALUES"); err != nil {
				return "", nil, err
			}
		}
	} else {
		if _, err := buf.WriteString(" ("); err != nil {
			return "", nil, err
		}

		if err := statement.dialect.Quoter().JoinWrite(buf.Builder, append(colNames, exprs.ColNames()...), ","); err != nil {
			return "", nil, err
		}

		if _, err := buf.WriteString(")"); err != nil {
			return "", nil, err
		}
		if err := statement.writeInsertOutput(buf.Builder, table); err != nil {
			return "", nil, err
		}

		if statement.Conds().IsValid() {
			if err := statement.writeStrings(" SELECT ")(buf); err != nil {
				return "", nil, err
			}

			if err := statement.WriteArgs(buf, args); err != nil {
				return "", nil, err
			}

			if len(exprs) > 0 {
				if _, err := buf.WriteString(","); err != nil {
					return "", nil, err
				}
				if err := exprs.WriteArgs(buf); err != nil {
					return "", nil, err
				}
			}

			if _, err := buf.WriteString(" FROM "); err != nil {
				return "", nil, err
			}

			statement.dialect.Quoter().QuoteTo(buf.Builder, tableName)

			if err := statement.writeWhere(buf); err != nil {
				return "", nil, err
			}
		} else {
			if _, err := buf.WriteString(" VALUES ("); err != nil {
				return "", nil, err
			}

			if err := statement.WriteArgs(buf, args); err != nil {
				return "", nil, err
			}

			if len(exprs) > 0 {
				if _, err := buf.WriteString(","); err != nil {
					return "", nil, err
				}
			}

			if err := exprs.WriteArgs(buf); err != nil {
				return "", nil, err
			}

			if _, err := buf.WriteString(")"); err != nil {
				return "", nil, err
			}
		}
	}

	if len(table.AutoIncrement) > 0 && statement.dialect.URI().DBType == schemas.POSTGRES {
		if _, err := buf.WriteString(" RETURNING "); err != nil {
			return "", nil, err
		}
		statement.dialect.Quoter().QuoteTo(buf.Builder, table.AutoIncrement)
	}

	return buf.String(), buf.Args(), nil
}

// GenInsertMapSQL generates insert map SQL
func (statement *Statement) GenInsertMapSQL(columns []string, args []any) (string, []any, error) {
	var (
		buf       = builder.NewWriter()
		exprs     = statement.ExprColumns
		tableName = statement.TableName()
	)

	if _, err := fmt.Fprintf(buf, "INSERT INTO %s (", statement.quote(tableName)); err != nil {
		return "", nil, err
	}

	if err := statement.dialect.Quoter().JoinWrite(buf.Builder, append(columns, exprs.ColNames()...), ","); err != nil {
		return "", nil, err
	}

	// if insert where
	if statement.Conds().IsValid() {
		if _, err := buf.WriteString(") SELECT "); err != nil {
			return "", nil, err
		}

		if err := statement.WriteArgs(buf, args); err != nil {
			return "", nil, err
		}

		if len(exprs) > 0 {
			if _, err := buf.WriteString(","); err != nil {
				return "", nil, err
			}
			if err := exprs.WriteArgs(buf); err != nil {
				return "", nil, err
			}
		}

		if _, err := fmt.Fprintf(buf, " FROM %s WHERE ", statement.quote(tableName)); err != nil {
			return "", nil, err
		}

		if err := statement.Conds().WriteTo(buf); err != nil {
			return "", nil, err
		}
	} else {
		if _, err := buf.WriteString(") VALUES ("); err != nil {
			return "", nil, err
		}
		if err := statement.WriteArgs(buf, args); err != nil {
			return "", nil, err
		}

		if len(exprs) > 0 {
			if _, err := buf.WriteString(","); err != nil {
				return "", nil, err
			}
			if err := exprs.WriteArgs(buf); err != nil {
				return "", nil, err
			}
		}
		if _, err := buf.WriteString(")"); err != nil {
			return "", nil, err
		}
	}

	return buf.String(), buf.Args(), nil
}

func (statement *Statement) GenInsertMultipleMapSQL(columns []string, argss [][]any) (string, []any, error) {
	var (
		buf       = builder.NewWriter()
		exprs     = statement.ExprColumns
		tableName = statement.TableName()
	)

	if _, err := fmt.Fprintf(buf, "INSERT INTO %s (", statement.quote(tableName)); err != nil {
		return "", nil, err
	}

	if err := statement.dialect.Quoter().JoinWrite(buf.Builder, append(columns, exprs.ColNames()...), ","); err != nil {
		return "", nil, err
	}

	// if insert where
	if statement.Conds().IsValid() {
		return "", nil, errors.New("batch insert don't support with where")
	}

	if _, err := buf.WriteString(") VALUES "); err != nil {
		return "", nil, err
	}
	for i, args := range argss {
		if _, err := buf.WriteString("("); err != nil {
			return "", nil, err
		}
		if err := statement.WriteArgs(buf, args); err != nil {
			return "", nil, err
		}

		if len(exprs) > 0 {
			if _, err := buf.WriteString(","); err != nil {
				return "", nil, err
			}
			if err := exprs.WriteArgs(buf); err != nil {
				return "", nil, err
			}
		}
		if _, err := buf.WriteString(")"); err != nil {
			return "", nil, err
		}
		if i < len(argss)-1 {
			if _, err := buf.WriteString(","); err != nil {
				return "", nil, err
			}
		}
	}

	return buf.String(), buf.Args(), nil
}

func (statement *Statement) writeColumns(w *builder.BytesWriter, slice []string) error {
	for i, s := range slice {
		if i > 0 {
			if _, err := fmt.Fprint(w, ","); err != nil {
				return err
			}
		}
		statement.dialect.Quoter().QuoteTo(w.Builder, s)
	}
	return nil
}

func (statement *Statement) WriteInsertMultiple(w *builder.BytesWriter, tableName string, colNames, colMultiPlaces []string) error {
	return statement.plainWriteInsertMultiple(w, tableName, colNames, colMultiPlaces)
}

func (statement *Statement) plainWriteInsertMultiple(w *builder.BytesWriter, tableName string, colNames, colMultiPlaces []string) error {
	if _, err := fmt.Fprint(w, "INSERT INTO "); err != nil {
		return err
	}
	statement.dialect.Quoter().QuoteTo(w.Builder, tableName)
	if _, err := fmt.Fprint(w, " ("); err != nil {
		return err
	}
	if err := statement.writeColumns(w, colNames); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, ") VALUES ("); err != nil {
		return err
	}
	for i, cols := range colMultiPlaces {
		if _, err := fmt.Fprint(w, cols, ")"); err != nil {
			return err
		}
		if i < len(colMultiPlaces)-1 {
			if _, err := fmt.Fprint(w, ",("); err != nil {
				return err
			}
		}
	}
	return nil
}
