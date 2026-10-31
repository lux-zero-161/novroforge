// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schemas

import (
	"database/sql"
	"math/big"
	"reflect"
	"strings"
	"time"
)

// DBType represents a database type
type DBType string

// enumerates all database types
const (
	POSTGRES DBType = "postgres"
	SQLITE   DBType = "sqlite3"
	MYSQL    DBType = "mysql"
)

// SQLType represents SQL types
type SQLType struct {
	Name           string
	DefaultLength  int64
	DefaultLength2 int64
}

// enumerates all columns types
const (
	UNKNOW_TYPE = iota
	TEXT_TYPE
	BLOB_TYPE
	TIME_TYPE
	NUMERIC_TYPE
	ARRAY_TYPE
	BOOL_TYPE
)

// IsType reutrns ture if the column type is the same as the parameter
func (s *SQLType) IsType(st int) bool {
	if t, ok := SqlTypes[s.Name]; ok && t == st {
		return true
	}
	return false
}

// IsText returns true if column is a text type
func (s *SQLType) IsText() bool {
	return s.IsType(TEXT_TYPE)
}

// IsBlob returns true if column is a binary type
func (s *SQLType) IsBlob() bool {
	return s.IsType(BLOB_TYPE)
}

// IsTime returns true if column is a time type
func (s *SQLType) IsTime() bool {
	return s.IsType(TIME_TYPE)
}

// IsBool returns true if column is a boolean type
func (s *SQLType) IsBool() bool {
	return s.IsType(BOOL_TYPE)
}

// IsNumeric returns true if column is a numeric type
func (s *SQLType) IsNumeric() bool {
	return s.IsType(NUMERIC_TYPE)
}

// IsArray returns true if column is an array type
func (s *SQLType) IsArray() bool {
	return s.IsType(ARRAY_TYPE)
}

// IsJson returns true if column is an array type
func (s *SQLType) IsJson() bool {
	return s.Name == Json || s.Name == Jsonb
}

// IsXML returns true if column is an xml type
func (s *SQLType) IsXML() bool {
	return s.Name == XML
}

// enumerates all the database column types
var (
	Bit               = "BIT"
	UnsignedBit       = "UNSIGNED BIT"
	TinyInt           = "TINYINT"
	UnsignedTinyInt   = "UNSIGNED TINYINT"
	SmallInt          = "SMALLINT"
	UnsignedSmallInt  = "UNSIGNED SMALLINT"
	MediumInt         = "MEDIUMINT"
	UnsignedMediumInt = "UNSIGNED MEDIUMINT"
	Int               = "INT"
	UnsignedInt       = "UNSIGNED INT"
	Integer           = "INTEGER"
	BigInt            = "BIGINT"
	UnsignedBigInt    = "UNSIGNED BIGINT"
	Number            = "NUMBER"

	Enum = "ENUM"
	Set  = "SET"

	Char             = "CHAR"
	Varchar          = "VARCHAR"
	VARCHAR2         = "VARCHAR2"
	NChar            = "NCHAR"
	NVarchar         = "NVARCHAR"
	TinyText         = "TINYTEXT"
	Text             = "TEXT"
	NText            = "NTEXT"
	Clob             = "CLOB"
	MediumText       = "MEDIUMTEXT"
	LongText         = "LONGTEXT"
	Uuid             = "UUID"
	UniqueIdentifier = "UNIQUEIDENTIFIER"
	SysName          = "SYSNAME"

	Date          = "DATE"
	DateTime      = "DATETIME"
	SmallDateTime = "SMALLDATETIME"
	Time          = "TIME"
	TimeStamp     = "TIMESTAMP"
	TimeStampz    = "TIMESTAMPZ"
	Year          = "YEAR"

	Decimal    = "DECIMAL"
	Numeric    = "NUMERIC"
	Money      = "MONEY"
	SmallMoney = "SMALLMONEY"

	Real          = "REAL"
	Float         = "FLOAT"
	UnsignedFloat = "UNSIGNED FLOAT"
	Double        = "DOUBLE"

	Binary     = "BINARY"
	VarBinary  = "VARBINARY"
	TinyBlob   = "TINYBLOB"
	Blob       = "BLOB"
	MediumBlob = "MEDIUMBLOB"
	LongBlob   = "LONGBLOB"
	Bytea      = "BYTEA"

	Bool    = "BOOL"
	Boolean = "BOOLEAN"

	Serial    = "SERIAL"
	BigSerial = "BIGSERIAL"

	Json  = "JSON"
	Jsonb = "JSONB"

	XML   = "XML"
	Array = "ARRAY"

	SqlTypes = map[string]int{
		Bit:               NUMERIC_TYPE,
		UnsignedBit:       NUMERIC_TYPE,
		TinyInt:           NUMERIC_TYPE,
		UnsignedTinyInt:   NUMERIC_TYPE,
		SmallInt:          NUMERIC_TYPE,
		UnsignedSmallInt:  NUMERIC_TYPE,
		MediumInt:         NUMERIC_TYPE,
		UnsignedMediumInt: NUMERIC_TYPE,
		Int:               NUMERIC_TYPE,
		UnsignedInt:       NUMERIC_TYPE,
		Integer:           NUMERIC_TYPE,
		BigInt:            NUMERIC_TYPE,
		UnsignedBigInt:    NUMERIC_TYPE,
		Number:            NUMERIC_TYPE,

		Enum:  TEXT_TYPE,
		Set:   TEXT_TYPE,
		Json:  TEXT_TYPE,
		Jsonb: TEXT_TYPE,

		XML: TEXT_TYPE,

		Char:       TEXT_TYPE,
		NChar:      TEXT_TYPE,
		Varchar:    TEXT_TYPE,
		VARCHAR2:   TEXT_TYPE,
		NVarchar:   TEXT_TYPE,
		TinyText:   TEXT_TYPE,
		Text:       TEXT_TYPE,
		NText:      TEXT_TYPE,
		MediumText: TEXT_TYPE,
		LongText:   TEXT_TYPE,
		Uuid:       TEXT_TYPE,
		Clob:       TEXT_TYPE,
		SysName:    TEXT_TYPE,

		Date:          TIME_TYPE,
		DateTime:      TIME_TYPE,
		Time:          TIME_TYPE,
		TimeStamp:     TIME_TYPE,
		TimeStampz:    TIME_TYPE,
		SmallDateTime: TIME_TYPE,
		Year:          TIME_TYPE,

		Decimal:       NUMERIC_TYPE,
		Numeric:       NUMERIC_TYPE,
		Real:          NUMERIC_TYPE,
		Float:         NUMERIC_TYPE,
		UnsignedFloat: NUMERIC_TYPE,
		Double:        NUMERIC_TYPE,
		Money:         NUMERIC_TYPE,
		SmallMoney:    NUMERIC_TYPE,

		Binary:    BLOB_TYPE,
		VarBinary: BLOB_TYPE,

		TinyBlob:         BLOB_TYPE,
		Blob:             BLOB_TYPE,
		MediumBlob:       BLOB_TYPE,
		LongBlob:         BLOB_TYPE,
		Bytea:            BLOB_TYPE,
		UniqueIdentifier: BLOB_TYPE,

		Bool:    BOOL_TYPE,
		Boolean: BOOL_TYPE,

		Serial:    NUMERIC_TYPE,
		BigSerial: NUMERIC_TYPE,

		"INT8": NUMERIC_TYPE,

		Array: ARRAY_TYPE,
	}
)

// enumerates all types
var (
	IntType   = reflect.TypeFor[int]()
	Int8Type  = reflect.TypeFor[int8]()
	Int16Type = reflect.TypeFor[int16]()
	Int32Type = reflect.TypeFor[int32]()
	Int64Type = reflect.TypeFor[int64]()

	UintType   = reflect.TypeFor[uint]()
	Uint8Type  = reflect.TypeFor[uint8]()
	Uint16Type = reflect.TypeFor[uint16]()
	Uint32Type = reflect.TypeFor[uint32]()
	Uint64Type = reflect.TypeFor[uint64]()

	Float32Type = reflect.TypeFor[float32]()
	Float64Type = reflect.TypeFor[float64]()

	Complex64Type  = reflect.TypeFor[complex64]()
	Complex128Type = reflect.TypeFor[complex128]()

	StringType = reflect.TypeFor[string]()
	BoolType   = reflect.TypeFor[bool]()
	ByteType   = reflect.TypeFor[byte]()
	BytesType  = reflect.SliceOf(ByteType)

	TimeType        = reflect.TypeFor[time.Time]()
	BigFloatType    = reflect.TypeFor[big.Float]()
	NullFloat64Type = reflect.TypeFor[sql.NullFloat64]()
	NullStringType  = reflect.TypeFor[sql.NullString]()
	NullInt32Type   = reflect.TypeFor[sql.NullInt32]()
	NullInt64Type   = reflect.TypeFor[sql.NullInt64]()
	NullBoolType    = reflect.TypeFor[sql.NullBool]()
)

type SQLTypeDelegator interface {
	// If implemented on a type, allows that type to be used as a column in a bean definition but be reflected in the
	// SQL column by a different type, the delegated type. The returned type from DelegateSQLType() will be mapped to a
	// SQLType through xorm's normal mapping procedure.
	DelegateSQLType() reflect.Type
}

// Type2SQLType generate SQLType acorrding Go's type
func Type2SQLType(t reflect.Type) (st SQLType) {
	val := reflect.New(t)
	if val.CanInterface() {
		valIface := val.Interface()
		introspect, ok := valIface.(SQLTypeDelegator)
		if ok {
			return Type2SQLType(introspect.DelegateSQLType())
		}
	}

	switch k := t.Kind(); k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		st = SQLType{Int, 0, 0}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		st = SQLType{UnsignedInt, 0, 0}
	case reflect.Int64:
		st = SQLType{BigInt, 0, 0}
	case reflect.Uint64:
		st = SQLType{UnsignedBigInt, 0, 0}
	case reflect.Float32:
		st = SQLType{Float, 0, 0}
	case reflect.Float64:
		st = SQLType{Double, 0, 0}
	case reflect.Complex64, reflect.Complex128:
		st = SQLType{Varchar, 64, 0}
	case reflect.Array, reflect.Slice, reflect.Map:
		if t.Elem() == ByteType {
			st = SQLType{Blob, 0, 0}
		} else {
			st = SQLType{Text, 0, 0}
		}
	case reflect.Bool:
		st = SQLType{Bool, 0, 0}
	case reflect.String:
		st = SQLType{Varchar, 255, 0}
	case reflect.Struct:
		if t.ConvertibleTo(TimeType) {
			st = SQLType{DateTime, 0, 0}
		} else if t.ConvertibleTo(NullFloat64Type) {
			st = SQLType{Double, 0, 0}
		} else if t.ConvertibleTo(NullStringType) {
			st = SQLType{Varchar, 255, 0}
		} else if t.ConvertibleTo(NullInt32Type) {
			st = SQLType{Integer, 0, 0}
		} else if t.ConvertibleTo(NullInt64Type) {
			st = SQLType{BigInt, 0, 0}
		} else if t.ConvertibleTo(NullBoolType) {
			st = SQLType{Boolean, 0, 0}
		} else {
			// TODO need to handle association struct
			st = SQLType{Text, 0, 0}
		}
	case reflect.Pointer:
		st = Type2SQLType(t.Elem())
	default:
		st = SQLType{Text, 0, 0}
	}
	return st
}

// SQLType2Type convert default sql type change to go types
func SQLType2Type(st SQLType) reflect.Type {
	name := strings.ToUpper(st.Name)
	switch name {
	case Bit, TinyInt, SmallInt, MediumInt, Int, Integer, Serial:
		return IntType
	case BigInt, BigSerial:
		return Int64Type
	case UnsignedBit, UnsignedTinyInt, UnsignedSmallInt, UnsignedMediumInt, UnsignedInt:
		return UintType
	case UnsignedBigInt:
		return Uint64Type
	case Float, Real:
		return Float32Type
	case Double:
		return Float64Type
	case Char, NChar, Varchar, NVarchar, TinyText, Text, NText, MediumText, LongText, Enum, Set, Uuid, Clob, SysName:
		return StringType
	case TinyBlob, Blob, LongBlob, Bytea, Binary, MediumBlob, VarBinary, UniqueIdentifier:
		return BytesType
	case Bool:
		return BoolType
	case DateTime, Date, Time, TimeStamp, TimeStampz, SmallDateTime, Year:
		return TimeType
	case Decimal, Numeric, Money, SmallMoney:
		return StringType
	default:
		return StringType
	}
}

// SQLTypeName returns sql type name
func SQLTypeName(tp string) string {
	fields := strings.Split(tp, "(")
	return fields[0]
}
