package mysqlconn

import (
	"fmt"

	"github.com/youtube/vitess/go/sqltypes"

	querypb "github.com/youtube/vitess/go/vt/proto/query"
)

// This file provides a few utility variables and methods, mostly for tests.
// The assumptions made about the types of fields and data returned
// by MySQl are validated in schema_test.go. This way all tests
// can use these variables and methods to simulate a MySQL server
// (using fakesqldb/ package for instance) and still be guaranteed correct
// data.

// DescribeTableFields contains the fields returned by a
// 'describe <table>' command. They are validated by the testDescribeTable
// test.
var DescribeTableFields = []*querypb.Field{
	{
		Name:         "Field",
		Type:         querypb.Type_VARCHAR,
		Table:        "COLUMNS",
		OrgTable:     "COLUMNS",
		Database:     "information_schema",
		OrgName:      "COLUMN_NAME",
		ColumnLength: 192,
		Charset:      33,
		Flags:        1,
	},
	{
		Name:         "Type",
		Type:         querypb.Type_TEXT,
		Table:        "COLUMNS",
		OrgTable:     "COLUMNS",
		Database:     "information_schema",
		OrgName:      "COLUMN_TYPE",
		ColumnLength: 589815,
		Charset:      33,
		Flags:        17,
	},
	{
		Name:         "Null",
		Type:         querypb.Type_VARCHAR,
		Table:        "COLUMNS",
		OrgTable:     "COLUMNS",
		Database:     "information_schema",
		OrgName:      "IS_NULLABLE",
		ColumnLength: 9,
		Charset:      33,
		Flags:        1,
	},
	{
		Name:         "Key",
		Type:         querypb.Type_VARCHAR,
		Table:        "COLUMNS",
		OrgTable:     "COLUMNS",
		Database:     "information_schema",
		OrgName:      "COLUMN_KEY",
		ColumnLength: 9,
		Charset:      33,
		Flags:        1,
	},
	{
		Name:         "Default",
		Type:         querypb.Type_TEXT,
		Table:        "COLUMNS",
		OrgTable:     "COLUMNS",
		Database:     "information_schema",
		OrgName:      "COLUMN_DEFAULT",
		ColumnLength: 589815,
		Charset:      33,
		Flags:        16,
	},
	{
		Name:         "Extra",
		Type:         querypb.Type_VARCHAR,
		Table:        "COLUMNS",
		OrgTable:     "COLUMNS",
		Database:     "information_schema",
		OrgName:      "EXTRA",
		ColumnLength: 90,
		Charset:      33,
		Flags:        1,
	},
}

// DescribeTableRow returns a row for a 'describe table' command.
// 'name' is the name of the field.
// 'type' is the type of the field. Something like:
//   'int(11)' for 'int'
//   'int(10) unsigned' for 'int unsigned'
//   'bigint(20)' for 'bigint'
//   'bigint(20) unsigned' for 'bigint unsigned'
//   'varchar(128)'
// 'null' is true if the field can be NULL.
// 'key' is either:
//    - 'PRI' if part of the primary key. If not:
//    - 'UNI' if part of a unique index. If not:
//    - 'MUL' if part of a non-unique index. If not:
//    - empty if part of no key / index.
// 'def' is the default value for the field. Empty if NULL default.
func DescribeTableRow(name string, typ string, null bool, key string, def string) []sqltypes.Value {
	nullStr := "NO"
	if null {
		nullStr = "YES"
	}
	defCell := sqltypes.NULL
	if def != "" {
		defCell = sqltypes.MakeTrusted(sqltypes.Text, []byte(def))
	}
	return []sqltypes.Value{
		sqltypes.MakeTrusted(sqltypes.VarChar, []byte(name)),
		sqltypes.MakeTrusted(sqltypes.Text, []byte(typ)),
		sqltypes.MakeTrusted(sqltypes.VarChar, []byte(nullStr)),
		sqltypes.MakeTrusted(sqltypes.VarChar, []byte(key)),
		defCell,
		sqltypes.MakeTrusted(sqltypes.VarChar, []byte("")), //extra
	}
}

// ShowIndexFromTableFields contains the fields returned by a 'show
// index from <table>' command. They are validated by the
// testShowIndexFromTable test.
var ShowIndexFromTableFields = []*querypb.Field{
	{
		Name:         "Table",
		Type:         querypb.Type_VARCHAR,
		Table:        "STATISTICS",
		OrgTable:     "STATISTICS",
		Database:     "information_schema",
		OrgName:      "TABLE_NAME",
		ColumnLength: 192,
		Charset:      CharacterSetUtf8,
		Flags:        1,
	},
	{
		Name:         "Non_unique",
		Type:         querypb.Type_INT64,
		Table:        "STATISTICS",
		OrgTable:     "STATISTICS",
		Database:     "information_schema",
		OrgName:      "NON_UNIQUE",
		ColumnLength: 1,
		Charset:      CharacterSetBinary,
		Flags:        1,
	},
	{
		Name:         "Key_name",
		Type:         querypb.Type_VARCHAR,
		Table:        "STATISTICS",
		OrgTable:     "STATISTICS",
		Database:     "information_schema",
		OrgName:      "INDEX_NAME",
		ColumnLength: 192,
		Charset:      33,
		Flags:        1,
	},
	{
		Name:         "Seq_in_index",
		Type:         querypb.Type_INT64,
		Table:        "STATISTICS",
		OrgTable:     "STATISTICS",
		Database:     "information_schema",
		OrgName:      "SEQ_IN_INDEX",
		ColumnLength: 2,
		Charset:      CharacterSetBinary,
		Flags:        1,
	},
	{
		Name:         "Column_name",
		Type:         querypb.Type_VARCHAR,
		Table:        "STATISTICS",
		OrgTable:     "STATISTICS",
		Database:     "information_schema",
		OrgName:      "COLUMN_NAME",
		ColumnLength: 192,
		Charset:      33,
		Flags:        1,
	},
	{
		Name:         "Collation",
		Type:         querypb.Type_VARCHAR,
		Table:        "STATISTICS",
		OrgTable:     "STATISTICS",
		Database:     "information_schema",
		OrgName:      "COLLATION",
		ColumnLength: 3,
		Charset:      33,
	},
	{
		Name:         "Cardinality",
		Type:         querypb.Type_INT64,
		Table:        "STATISTICS",
		OrgTable:     "STATISTICS",
		Database:     "information_schema",
		OrgName:      "CARDINALITY",
		ColumnLength: 21,
		Charset:      CharacterSetBinary,
	},
	{
		Name:         "Sub_part",
		Type:         querypb.Type_INT64,
		Table:        "STATISTICS",
		OrgTable:     "STATISTICS",
		Database:     "information_schema",
		OrgName:      "SUB_PART",
		ColumnLength: 3,
		Charset:      CharacterSetBinary,
	},
	{
		Name:         "Packed",
		Type:         querypb.Type_VARCHAR,
		Table:        "STATISTICS",
		OrgTable:     "STATISTICS",
		Database:     "information_schema",
		OrgName:      "PACKED",
		ColumnLength: 30,
		Charset:      33,
	},
	{
		Name:         "Null",
		Type:         querypb.Type_VARCHAR,
		Table:        "STATISTICS",
		OrgTable:     "STATISTICS",
		Database:     "information_schema",
		OrgName:      "NULLABLE",
		ColumnLength: 9,
		Charset:      33,
		Flags:        1,
	},
	{
		Name:         "Index_type",
		Type:         querypb.Type_VARCHAR,
		Table:        "STATISTICS",
		OrgTable:     "STATISTICS",
		Database:     "information_schema",
		OrgName:      "INDEX_TYPE",
		ColumnLength: 48,
		Charset:      33,
		Flags:        1,
	},
	{
		Name:         "Comment",
		Type:         querypb.Type_VARCHAR,
		Table:        "STATISTICS",
		OrgTable:     "STATISTICS",
		Database:     "information_schema",
		OrgName:      "COMMENT",
		ColumnLength: 48,
		Charset:      33,
	},
	{
		Name:         "Index_comment",
		Type:         querypb.Type_VARCHAR,
		Table:        "STATISTICS",
		OrgTable:     "STATISTICS",
		Database:     "information_schema",
		OrgName:      "INDEX_COMMENT",
		ColumnLength: 3072,
		Charset:      33,
		Flags:        1,
	},
}

// ShowIndexFromTableRow returns the fields from a 'show index from table'
// command.
// 'table' is the table name.
// 'unique' is true for unique indexes, false for non-unique indexes.
// 'keyName' is 'PRIMARY' for PKs, otherwise the name of the index.
// 'seqInIndex' is starting at 1 for first key in index.
// 'columnName' is the name of the column this index applies to.
// 'nullable' is true if this column can be null.
func ShowIndexFromTableRow(table string, unique bool, keyName string, seqInIndex int, columnName string, nullable bool) []sqltypes.Value {
	nonUnique := "1"
	if unique {
		nonUnique = "0"
	}
	nullableStr := ""
	if nullable {
		nullableStr = "YES"
	}
	return []sqltypes.Value{
		sqltypes.MakeTrusted(sqltypes.VarChar, []byte(table)),
		sqltypes.MakeTrusted(sqltypes.Int64, []byte(nonUnique)),
		sqltypes.MakeTrusted(sqltypes.VarChar, []byte(keyName)),
		sqltypes.MakeTrusted(sqltypes.Int64, []byte(fmt.Sprintf("%v", seqInIndex))),
		sqltypes.MakeTrusted(sqltypes.VarChar, []byte(columnName)),
		sqltypes.MakeTrusted(sqltypes.VarChar, []byte("A")), // Collation
		sqltypes.MakeTrusted(sqltypes.Int64, []byte("0")),   // Cardinality
		sqltypes.NULL,                                       // Sub_part
		sqltypes.NULL,                                       // Packed
		sqltypes.MakeTrusted(sqltypes.VarChar, []byte(nullableStr)),
		sqltypes.MakeTrusted(sqltypes.VarChar, []byte("BTREE")), // Index_type
		sqltypes.MakeTrusted(sqltypes.VarChar, []byte("")),      // Comment
		sqltypes.MakeTrusted(sqltypes.VarChar, []byte("")),      // Index_comment
	}
}
