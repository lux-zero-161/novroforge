// Copyright 2015 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dialects

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"novroforge/xorm/xorm/core"
	"novroforge/xorm/xorm/schemas"
)

var (
	sqlite3ReservedWords = map[string]bool{
		"ABORT":             true,
		"ACTION":            true,
		"ADD":               true,
		"AFTER":             true,
		"ALL":               true,
		"ALTER":             true,
		"ANALYZE":           true,
		"AND":               true,
		"AS":                true,
		"ASC":               true,
		"ATTACH":            true,
		"AUTOINCREMENT":     true,
		"BEFORE":            true,
		"BEGIN":             true,
		"BETWEEN":           true,
		"BY":                true,
		"CASCADE":           true,
		"CASE":              true,
		"CAST":              true,
		"CHECK":             true,
		"COLLATE":           true,
		"COLUMN":            true,
		"COMMIT":            true,
		"CONFLICT":          true,
		"CONSTRAINT":        true,
		"CREATE":            true,
		"CROSS":             true,
		"CURRENT_DATE":      true,
		"CURRENT_TIME":      true,
		"CURRENT_TIMESTAMP": true,
		"DATABASE":          true,
		"DEFAULT":           true,
		"DEFERRABLE":        true,
		"DEFERRED":          true,
		"DELETE":            true,
		"DESC":              true,
		"DETACH":            true,
		"DISTINCT":          true,
		"DROP":              true,
		"EACH":              true,
		"ELSE":              true,
		"END":               true,
		"ESCAPE":            true,
		"EXCEPT":            true,
		"EXCLUSIVE":         true,
		"EXISTS":            true,
		"EXPLAIN":           true,
		"FAIL":              true,
		"FOR":               true,
		"FOREIGN":           true,
		"FROM":              true,
		"FULL":              true,
		"GLOB":              true,
		"GROUP":             true,
		"HAVING":            true,
		"IF":                true,
		"IGNORE":            true,
		"IMMEDIATE":         true,
		"IN":                true,
		"INDEX":             true,
		"INDEXED":           true,
		"INITIALLY":         true,
		"INNER":             true,
		"INSERT":            true,
		"INSTEAD":           true,
		"INTERSECT":         true,
		"INTO":              true,
		"IS":                true,
		"ISNULL":            true,
		"JOIN":              true,
		"KEY":               true,
		"LEFT":              true,
		"LIKE":              true,
		"LIMIT":             true,
		"MATCH":             true,
		"NATURAL":           true,
		"NO":                true,
		"NOT":               true,
		"NOTNULL":           true,
		"NULL":              true,
		"OF":                true,
		"OFFSET":            true,
		"ON":                true,
		"OR":                true,
		"ORDER":             true,
		"OUTER":             true,
		"PLAN":              true,
		"PRAGMA":            true,
		"PRIMARY":           true,
		"QUERY":             true,
		"RAISE":             true,
		"RECURSIVE":         true,
		"REFERENCES":        true,
		"REGEXP":            true,
		"REINDEX":           true,
		"RELEASE":           true,
		"RENAME":            true,
		"REPLACE":           true,
		"RESTRICT":          true,
		"RIGHT":             true,
		"ROLLBACK":          true,
		"ROW":               true,
		"SAVEPOINT":         true,
		"SELECT":            true,
		"SET":               true,
		"TABLE":             true,
		"TEMP":              true,
		"TEMPORARY":         true,
		"THEN":              true,
		"TO":                true,
		"TRANSACTI":         true,
		"TRIGGER":           true,
		"UNION":             true,
		"UNIQUE":            true,
		"UPDATE":            true,
		"USING":             true,
		"VACUUM":            true,
		"VALUES":            true,
		"VIEW":              true,
		"VIRTUAL":           true,
		"WHEN":              true,
		"WHERE":             true,
		"WITH":              true,
		"WITHOUT":           true,
	}

	sqlite3Quoter = schemas.Quoter{
		Prefix:     '`',
		Suffix:     '`',
		IsReserved: schemas.AlwaysReserve,
	}
)

type sqlite3 struct {
	Base
}

func (db *sqlite3) Init(uri *URI) error {
	db.quoter = sqlite3Quoter
	return db.Base.Init(db, uri)
}

func (db *sqlite3) Version(ctx context.Context, queryer core.Queryer) (*schemas.Version, error) {
	rows, err := queryer.QueryContext(ctx, "SELECT sqlite_version()")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var version string
	if !rows.Next() {
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		return nil, errors.New("unknow version")
	}

	if err := rows.Scan(&version); err != nil {
		return nil, err
	}
	return &schemas.Version{
		Number:  version,
		Edition: "sqlite",
	}, nil
}

func (db *sqlite3) Features() *DialectFeatures {
	return &DialectFeatures{
		AutoincrMode: IncrAutoincrMode,
	}
}

func (db *sqlite3) SetQuotePolicy(quotePolicy QuotePolicy) {
	switch quotePolicy {
	case QuotePolicyNone:
		q := sqlite3Quoter
		q.IsReserved = schemas.AlwaysNoReserve
		db.quoter = q
	case QuotePolicyReserved:
		q := sqlite3Quoter
		q.IsReserved = db.IsReserved
		db.quoter = q
	case QuotePolicyAlways:
		fallthrough
	default:
		db.quoter = sqlite3Quoter
	}
}

func (db *sqlite3) SQLType(c *schemas.Column) string {
	switch t := c.SQLType.Name; t {
	case schemas.Bool:
		switch c.Default {
		case "true":
			c.Default = "1"
		case "false":
			c.Default = "0"
		}
		return schemas.Integer
	case schemas.Date, schemas.DateTime, schemas.TimeStamp, schemas.Time:
		return schemas.DateTime
	case schemas.TimeStampz:
		return schemas.Text
	case schemas.Char, schemas.Varchar, schemas.NVarchar, schemas.TinyText,
		schemas.Text, schemas.MediumText, schemas.LongText, schemas.Json:
		return schemas.Text
	case schemas.Bit, schemas.TinyInt, schemas.UnsignedTinyInt, schemas.SmallInt,
		schemas.UnsignedSmallInt, schemas.MediumInt, schemas.Int, schemas.UnsignedInt,
		schemas.BigInt, schemas.UnsignedBigInt, schemas.Integer:
		return schemas.Integer
	case schemas.Float, schemas.Double, schemas.Real:
		return schemas.Real
	case schemas.Decimal, schemas.Numeric:
		return schemas.Numeric
	case schemas.TinyBlob, schemas.Blob, schemas.MediumBlob, schemas.LongBlob, schemas.Bytea, schemas.Binary, schemas.VarBinary:
		return schemas.Blob
	case schemas.Serial, schemas.BigSerial:
		c.IsPrimaryKey = true
		c.IsAutoIncrement = true
		c.Nullable = false
		return schemas.Integer
	default:
		return t
	}
}

func (db *sqlite3) ColumnTypeKind(t string) int {
	switch strings.ToUpper(t) {
	case "DATETIME":
		return schemas.TIME_TYPE
	case "TEXT":
		return schemas.TEXT_TYPE
	case "INTEGER", "REAL", "NUMERIC", "DECIMAL":
		return schemas.NUMERIC_TYPE
	case "BLOB":
		return schemas.BLOB_TYPE
	default:
		return schemas.UNKNOW_TYPE
	}
}

func (db *sqlite3) IsReserved(name string) bool {
	_, ok := sqlite3ReservedWords[strings.ToUpper(name)]
	return ok
}

func (db *sqlite3) AutoIncrStr() string {
	return "AUTOINCREMENT"
}

func (db *sqlite3) IndexCheckSQL(tableName, idxName string) (string, []any) {
	args := []any{idxName}
	return "SELECT name FROM sqlite_master WHERE type='index' and name = ?", args
}

func (db *sqlite3) IsTableExist(ctx context.Context, queryer core.Queryer, tableName string) (bool, error) {
	return db.HasRecords(ctx, queryer, "SELECT name FROM sqlite_master WHERE type='table' and name = ?", tableName)
}

func (db *sqlite3) DropIndexSQL(tableName string, index *schemas.Index) string {
	// var unique string
	idxName := index.Name

	if !strings.HasPrefix(idxName, "UQE_") &&
		!strings.HasPrefix(idxName, "IDX_") {
		if index.Type == schemas.UniqueType {
			idxName = fmt.Sprintf("UQE_%v_%v", tableName, index.Name)
		} else {
			idxName = fmt.Sprintf("IDX_%v_%v", tableName, index.Name)
		}
	}
	return fmt.Sprintf("DROP INDEX %v", db.Quoter().Quote(idxName))
}

func (db *sqlite3) IsColumnExist(ctx context.Context, queryer core.Queryer, tableName, colName string) (bool, error) {
	query := "SELECT * FROM " + tableName + " LIMIT 0"
	rows, err := queryer.QueryContext(ctx, query)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return false, err
	}

	for _, col := range cols {
		if strings.EqualFold(col, colName) {
			return true, nil
		}
	}

	return false, nil
}

// splitColStr splits a sqlite col strings as fields
func splitColStr(colStr string) []string {
	colStr = strings.TrimSpace(colStr)
	results := make([]string, 0, 10)
	var lastIdx int
	var hasC, hasQuote bool
	for i, c := range colStr {
		if unicode.IsSpace(c) && !hasQuote {
			if hasC {
				results = append(results, colStr[lastIdx:i])
				hasC = false
			}
		} else {
			if c == '\'' {
				hasQuote = !hasQuote
			}
			if !hasC {
				lastIdx = i
			}
			hasC = true
			if i == len(colStr)-1 {
				results = append(results, colStr[lastIdx:i+1])
			}
		}
	}
	return results
}

func (db *sqlite3) GetColumns(ctx context.Context, queryer core.Queryer, tableName string) ([]string, map[string]*schemas.Column, error) {
	args := []any{tableName}
	s := "SELECT `name`, `type`, `notnull`, `dflt_value`, `pk` FROM pragma_table_info(?)"
	rows, err := queryer.QueryContext(ctx, s, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	cols := make(map[string]*schemas.Column)
	colSeq := make([]string, 0)

	for rows.Next() {
		var name string
		var dataType string
		var notNull bool
		var defaultValue *string
		var primaryKeyIndex int // either zero for columns that are not part of the primary key, or the 1-based index of the column within the primary key

		err = rows.Scan(&name, &dataType, &notNull, &defaultValue, &primaryKeyIndex)
		if err != nil {
			return nil, nil, err
		}

		col := &schemas.Column{
			Name:         name,
			Indexes:      make(map[string]int),
			IsPrimaryKey: primaryKeyIndex != 0,
			Nullable:     !notNull,
			SQLType:      schemas.SQLType{Name: dataType, DefaultLength: 0, DefaultLength2: 0},
		}

		if defaultValue == nil {
			col.DefaultIsEmpty = true
		} else {
			col.DefaultIsEmpty = false
			col.Default = *defaultValue
		}

		cols[col.Name] = col
		colSeq = append(colSeq, col.Name)
	}
	if rows.Err() != nil {
		return nil, nil, rows.Err()
	}

	// `PRAGMA table_info()` doesn't include information about whether a column is autoincrementing, which is part of
	// our schema inspection structure.  Autoincrement can only be set on a primary key, so we make a best guess search
	// of the table for the `AUTOINCREMENT` keyword and then flag the primary key column if present:
	if autoIncrement, err := hasAutoIncrement(ctx, queryer, tableName); err != nil {
		return nil, nil, fmt.Errorf("hasAutoIncrement: %w", err)
	} else if autoIncrement {
		for _, c := range cols {
			if c.IsPrimaryKey {
				c.IsAutoIncrement = true
			}
		}
	}

	return colSeq, cols, nil
}

var autoincrementRegex = regexp.MustCompile(`(?i)\WAUTOINCREMENT\W`)

func hasAutoIncrement(ctx context.Context, queryer core.Queryer, tableName string) (bool, error) {
	args := []any{tableName}
	s := "SELECT sql FROM sqlite_master WHERE type='table' and name = ?"
	rows, err := queryer.QueryContext(ctx, s, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var tableDefinition string
	if rows.Next() {
		err = rows.Scan(&tableDefinition)
		if err != nil {
			return false, err
		}
	}
	if rows.Err() != nil {
		return false, rows.Err()
	}
	if tableDefinition == "" {
		return false, errors.New("no table found in sqlite_master named " + tableName)
	}

	return autoincrementRegex.MatchString(tableDefinition), nil
}

func (db *sqlite3) GetTables(ctx context.Context, queryer core.Queryer) ([]*schemas.Table, error) {
	args := []any{}
	s := "SELECT name FROM sqlite_master WHERE type='table'"

	rows, err := queryer.QueryContext(ctx, s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]*schemas.Table, 0)
	for rows.Next() {
		table := schemas.NewEmptyTable()
		err = rows.Scan(&table.Name)
		if err != nil {
			return nil, err
		}
		if table.Name == "sqlite_sequence" {
			continue
		}
		tables = append(tables, table)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return tables, nil
}

func (db *sqlite3) GetIndexes(ctx context.Context, queryer core.Queryer, tableName string) (map[string]*schemas.Index, error) {
	foreignKeys, err := db.dialect.GetForeignKeys(ctx, queryer, tableName)
	if err != nil {
		return nil, err
	}

	args := []any{tableName}
	s := "SELECT il.`name`, il.`unique`, il.`origin`, ii.`name` as field_name from pragma_index_list(?) as il, pragma_index_info(il.name) as ii order by il.seq, ii.`seqno`"
	rows, err := queryer.QueryContext(ctx, s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]*schemas.Index)

	var currentIndex *schemas.Index
	for rows.Next() {
		// pragma_index_list:
		var indexName string // The name of the index.
		var isUnique bool    // "1" if the index is UNIQUE and "0" if not.
		var origin string    // "c" if the index was created by a CREATE INDEX statement, "u" if the index was created by a UNIQUE constraint, or "pk" if the index was created by a PRIMARY KEY constraint.
		// pragma_index_info:
		var fieldName *string // The name of the column being indexed. This columns is NULL if the column is the rowid or an expression.

		err = rows.Scan(&indexName, &isUnique, &origin, &fieldName)
		if err != nil {
			return nil, err
		}
		if origin == "pk" {
			// Skip primary keys as indexes
			continue
		}

		var xormIndexName string
		var isRegular bool
		if strings.HasPrefix(indexName, "IDX_"+tableName) || strings.HasPrefix(indexName, "UQE_"+tableName) {
			xormIndexName = indexName[5+len(tableName):]
			isRegular = true
		} else {
			xormIndexName = indexName
		}

		// For first row, and whenever we transition to a new index, initialize the index object itself:
		if currentIndex == nil || currentIndex.Name != xormIndexName {
			currentIndex = &schemas.Index{
				Name:      xormIndexName,
				IsRegular: isRegular,
				Cols:      []string{},
			}

			switch origin {
			case "c":
				currentIndex.Type = schemas.IndexType
			case "u":
				currentIndex.Type = schemas.UniqueType
			default:
				return nil, fmt.Errorf("unrecognized index origin: %q", origin)
			}
			if isUnique {
				currentIndex.Type = schemas.UniqueType
			}

			for _, fk := range foreignKeys {
				if currentIndex.Name == fmt.Sprintf("%s_fkey", fk.SourceFieldName) {
					currentIndex.IsForeignKey = true
					break
				}
			}

			indexes[currentIndex.Name] = currentIndex
		}

		if fieldName != nil {
			currentIndex.Cols = append(currentIndex.Cols, *fieldName)
		}
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return indexes, nil
}

func (db *sqlite3) Filters() []Filter {
	return []Filter{}
}

func (db *sqlite3) GetForeignKeys(ctx context.Context, queryer core.Queryer, tableName string) ([]schemas.ForeignKey, error) {
	// Can't use ? in PRAGMA statement
	s := "PRAGMA foreign_key_list('" + tableName + "')"

	rows, err := queryer.QueryContext(ctx, s)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// First pass: collect all foreign key info
	type fkInfo struct {
		id              int
		seq             int
		targetTableName string
		sourceFieldName string
		targetFieldName string
		onUpdate        string
		onDelete        string
	}

	var allFKs []fkInfo
	for rows.Next() {
		var id int
		var seq int
		var targetTableName string
		var sourceFieldName string
		var targetFieldName string
		var onUpdate string
		var onDelete string
		var match string

		// PRAGMA foreign_key_list returns: id, seq, table, from, to, on_update, on_delete, match
		err = rows.Scan(&id, &seq, &targetTableName, &sourceFieldName, &targetFieldName, &onUpdate, &onDelete, &match)
		if err != nil {
			return nil, err
		}

		allFKs = append(allFKs, fkInfo{
			id:              id,
			seq:             seq,
			targetTableName: targetTableName,
			sourceFieldName: sourceFieldName,
			targetFieldName: targetFieldName,
			onUpdate:        onUpdate,
			onDelete:        onDelete,
		})
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	foreignKeys := make([]schemas.ForeignKey, 0)
	processedIDs := make(map[int]bool)

	for _, fk := range allFKs {
		// Skip if already processed
		if processedIDs[fk.id] {
			continue
		}

		// Apply filtering logic:
		// - Only NO ACTION for delete and update
		// - Only single-column foreign keys
		if fk.onDelete == "NO ACTION" && fk.onUpdate == "NO ACTION" && fk.seq == 0 {
			// Check if this is a multi-column foreign key
			isMultiColumn := false
			for _, otherFK := range allFKs {
				if otherFK.id == fk.id && otherFK.seq > 0 {
					isMultiColumn = true
					break
				}
			}

			if !isMultiColumn {
				// SQLite FKs don't have names... but the source table and source field name should be unique.  Matches
				// pgsql's implementation for single-column FKs as well.
				constraintName := fmt.Sprintf("%s_%s_fkey", tableName, fk.sourceFieldName)

				foreignKeys = append(foreignKeys, schemas.ForeignKey{
					Name:            constraintName,
					SourceFieldName: fk.sourceFieldName,
					TargetTableName: fk.targetTableName,
					TargetFieldName: fk.targetFieldName,
				})
			}
		}

		processedIDs[fk.id] = true
	}

	return foreignKeys, nil
}

func (db *sqlite3) GetAlterTableSQL(originalTable *schemas.Table, tableName string, modifications ...schemas.TableModification) ([]AlterTableOp, error) {
	if len(modifications) == 0 {
		return []AlterTableOp{}, nil
	}

	statements := []AlterTableOp{}

	// Deep-copy schemas.Table so that we can apply modifications to a local copy.  Only deeply copies ForeignKeys for
	// now since that's all that is modified at this time.
	updatedTable := *originalTable
	updatedTable.ForeignKeys = slices.Clone(originalTable.ForeignKeys)

	for _, modification := range modifications {
		switch m := modification.(type) {
		case schemas.AddForeignKey:
			updatedTable.ForeignKeys = append(updatedTable.ForeignKeys, m.ForeignKey)
		case schemas.DropForeignKey:
			idx := slices.Index(updatedTable.ForeignKeys, m.ForeignKey)
			// If originalTable is based upon a bean that has already has the foreign key removed, then we won't find
			// it... which is fine because we would have just removed it anyway.
			if idx != -1 {
				updatedTable.ForeignKeys = slices.Delete(updatedTable.ForeignKeys, idx, idx+1)
			}
		default:
			return nil, fmt.Errorf("unknown modification type: %T", modification)
		}
	}

	// SQLite only supports a handful of table alterations -- renaming a table, renaming a column, adding a column, and
	// dropping a column.  For all other modifications, they document a specific workflow to create a new table and
	// rewrite the data: https://www.sqlite.org/lang_altertable.html#making_other_kinds_of_table_schema_changes So, here
	// we go...

	// 1. If foreign key constraints are enabled, disable them using PRAGMA foreign_keys=OFF.
	statements = append(statements, MkSQLOp("PRAGMA foreign_keys=OFF"))

	// 2. Start a transaction.
	statements = append(statements, AlterTableOp{OpType: TxStartOp})

	// X. Defer foreign key enforcement.
	//    This also works inside of transactions, and handles the case that the alter operation is run in an outer transaction,
	//    which prevents the PRAGMA foreign_keys=OFF from working.
	statements = append(statements, MkSQLOp("PRAGMA defer_foreign_keys=ON"))

	// 3. Remember the format of all indexes, triggers, and views associated with table X. This information will be
	// needed in step 8 below. One way to do this is to run a query like the following: SELECT type, sql FROM
	// sqlite_schema WHERE tbl_name='X'.
	//
	// implemented by the passed-in `originalTable`, so that this dialect function doesn't require a reference to an
	// engine.

	// 4. Use CREATE TABLE to construct a new table "new_X" that is in the desired revised format of table X. Make sure
	// that the name "new_X" does not collide with any existing table name, of course.
	temporaryTableName := fmt.Sprintf("__alter_%s", tableName)
	createTableSQL, _, err := db.dialect.CreateTableSQL(&updatedTable, temporaryTableName)
	if err != nil {
		return nil, err
	}
	statements = append(statements, MkSQLOp(createTableSQL))

	// 5. Transfer content from X into new_X using a statement like: INSERT INTO new_X SELECT ... FROM X.
	columnNames := make([]string, 0, len(originalTable.ColumnsSeq()))
	for _, colName := range originalTable.ColumnsSeq() {
		c := originalTable.GetColumn(colName)
		columnNames = append(columnNames, db.Quoter().Quote(c.Name))
	}
	columns := strings.Join(columnNames, ", ")
	var b strings.Builder
	b.WriteString("INSERT INTO ")
	b.WriteString(temporaryTableName)
	b.WriteString(" (")
	b.WriteString(columns)
	b.WriteString(") SELECT ")
	b.WriteString(columns)
	b.WriteString(" FROM ")
	b.WriteString(tableName)
	statements = append(statements, MkSQLOp(b.String()))

	// 6. Drop the old table X: DROP TABLE X.
	statements = append(statements, MkSQLOp(fmt.Sprintf("DROP TABLE %s", tableName)))

	// 7. Change the name of new_X to X using: ALTER TABLE new_X RENAME TO X.
	statements = append(statements, MkSQLOp(fmt.Sprintf("ALTER TABLE %s RENAME TO %s", temporaryTableName, tableName)))

	// 8. Use CREATE INDEX, CREATE TRIGGER, and CREATE VIEW to reconstruct indexes, triggers, and views associated with
	// table X. Perhaps use the old format of the triggers, indexes, and views saved from step 3 above as a guide,
	// making changes as appropriate for the alteration.
	for _, index := range updatedTable.Indexes {
		statements = append(statements, MkSQLOp(db.dialect.CreateIndexSQL(tableName, index)))
	}

	// 9. If any views refer to table X in a way that is affected by the schema change, then drop those views using DROP
	// VIEW and recreate them with whatever changes are necessary to accommodate the schema change using CREATE VIEW.
	//
	// This is not implemented here.  It is a documented requirement of using `AlterTable` on `Session`; the library
	// user must handle it.

	// X2. Turn foreign key deferment off again.
	//     Doing this inside of the same transactions cancels all pending FK failures, which did still accumulate and would be reported on COMMIT.
	statements = append(statements, MkSQLOp("PRAGMA defer_foreign_keys=OFF"))

	// 10. If foreign key constraints were originally enabled then run PRAGMA foreign_key_check to verify that the
	// schema change did not break any foreign key constraints.
	statements = append(statements, MkQueryOp(
		fmt.Sprintf("PRAGMA foreign_key_check(%s)", tableName),
		func(rows *sql.Rows) error {
			defer rows.Close()
			if rows.Next() {
				return ErrForeignKeyViolation
			}
			return rows.Err()
		},
	))

	// 11. Commit the transaction started in step 2.
	statements = append(statements, AlterTableOp{OpType: TxCommitOp})

	// 12. If foreign keys constraints were originally enabled, reenable them now.
	statements = append(statements, MkDeferredSQLOp("PRAGMA foreign_keys=ON"))

	return statements, nil
}

type sqlite3Driver struct {
	baseDriver
}

func (p *sqlite3Driver) Features() *DriverFeatures {
	return &DriverFeatures{
		SupportReturnInsertedID: true,
	}
}

func (p *sqlite3Driver) Parse(driverName, dataSourceName string) (*URI, string, error) {
	// In 2009 when foreign keys were added to SQLite, they chose to disable their enforcement by default because they
	// had previously accepted the syntax in table creation, and enforcing those foreign keys suddenly would be a
	// breaking change.  At the time of writing in 2025 this seems like an exceptionally unlikely problem to encounter
	// in a real-world scenario, so xorm enables foreign key support by default on SQLite DSNs.
	//
	// sqlite & sqlite3 allow enabling `PRAGMA foreign_keys = on` during connection by providing URL-like parameters,
	// ?_foreign_keys=true for github.com/mattn/go-sqlite3, and _pragma=foreign_keys(1) for modernc.org/sqlite.
	//
	// In the unlikely case where an xorm user would not like foreign keys to be enabled, they can provide an
	// ?_foreign_keys or ?_pragma=foreign_keys value in their DSN, and this driver will *not* add support itself.
	dsnHasFK := strings.Contains(dataSourceName, "_foreign_keys=") || strings.Contains(dataSourceName, "foreign_keys(")
	enableFK := "_foreign_keys=true&_pragma=foreign_keys(1)"

	databaseName := dataSourceName
	databaseName, _, found := strings.Cut(databaseName, "?")
	if !dsnHasFK {
		if found {
			dataSourceName = dataSourceName + "&" + enableFK
		} else {
			dataSourceName = dataSourceName + "?" + enableFK
		}
	}

	return &URI{DBType: schemas.SQLITE, DBName: databaseName}, dataSourceName, nil
}

func (p *sqlite3Driver) GenScanResult(colType string) (any, error) {
	switch colType {
	case "TEXT":
		var s sql.NullString
		return &s, nil
	case "INTEGER":
		var s sql.NullInt64
		return &s, nil
	case "DATETIME":
		var s sql.NullTime
		return &s, nil
	case "REAL":
		var s sql.NullFloat64
		return &s, nil
	case "NUMERIC", "DECIMAL":
		var s sql.NullString
		return &s, nil
	case "BLOB":
		var s sql.RawBytes
		return &s, nil
	default:
		var r sql.NullString
		return &r, nil
	}
}

func (p *sqlite3Driver) StandardizeError(err error) error {
	return sqliteStandardizeError(err)
}

type SQLiteError interface {
	Code() int
}

// Extract the error code field from a variety of SQLite error types, without linking those libraries.
// modernc.org/sqlite implements a `Code()` method on its error, and mattn/go-sqlite3 has an `ExtendedCode` field on its
// struct with no method to access; so we use a combination of interface inspection and reflection to find the common
// error code.
func getSqliteErrorCode(err error) (int, bool) {
	if err == nil {
		return 0, false
	}

	if sqliteError, ok := err.(SQLiteError); ok {
		return sqliteError.Code(), true
	}

	v := reflect.ValueOf(err)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return 0, false
	}

	// "github.com/mattn/go-sqlite3" has a `ExtendedCode ErrNoExtended` field on `Error`.
	codeField := v.FieldByName("ExtendedCode")
	if !codeField.IsValid() {
		return 0, false
	}
	if !codeField.CanInt() {
		return 0, false
	}

	return int(codeField.Int()), true
}

func sqliteStandardizeError(err error) error {
	if err == nil {
		return nil
	}

	errorNumber, ok := getSqliteErrorCode(err)
	if ok {
		switch errorNumber {
		case 787:
			return fmt.Errorf("%w: %w", ErrForeignKeyViolation, err)
		case 1555, 2067:
			return fmt.Errorf("%w: %w", ErrUniqueConstraintViolation, err)
		}
	}

	return err
}
