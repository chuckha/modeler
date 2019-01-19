package mysql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/chuckha/modeler/repr"
	"github.com/pkg/errors"
)

// Representer converts various data into a *Table representation.
type Representer struct{}

// RepresentationFromRows converts sql rows into a Table.
/*
	+-------------+-------------------+-------------+-----------+
	| COLUMN_NAME | COLUMN_DEFAULT    | IS_NULLABLE | DATA_TYPE |
	+-------------+-------------------+-------------+-----------+
	| id          | 0                 | YES         | int       |
	| user_id     | NULL              | YES         | text      |
	| ended       | 0                 | YES         | tinyint   |
	| preamble    | NULL              | YES         | text      |
	| created     | CURRENT_TIMESTAMP | NO          | timestamp |
	| updated     | CURRENT_TIMESTAMP | NO          | timestamp |
	+-------------+-------------------+-------------+-----------+
*/
func (r *Representer) RepresentationFromRows(rows *sql.Rows) (*repr.Table, error) {
	t := &repr.Table{
		Fields: make([]*repr.Field, 0),
	}
	defer rows.Close()
	for rows.Next() {
		f := &repr.Field{}
		var nullable string
		err := rows.Scan(&f.Name, &f.Default, &nullable, &f.DataType)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		f.Nullable = nullable == "YES"
		t.Fields = append(t.Fields, f)
	}
	return t, rows.Err()
}

// RepresentationFromInterface converts a model into a Table.
func (r *Representer) RepresentationFromInterface(model interface{}) *repr.Table {
	// TODO use this
	// table := &repr.Table{
	// 	Fields: make([]*repr.Field, 0),
	// }
	t := reflect.TypeOf(model).Elem()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		// Ignore unexported fields
		if sf.Name[0] == strings.ToLower(sf.Name)[0] {
			continue
		}
		sqlTagValue := sf.Tag.Get("sql")
		// Ignore unmarked fields
		if sqlTagValue == "" {
			continue
		}
		f := &repr.Field{}
		col := newCol(sf.Type)
		for _, tag := range strings.Split(sqlTagValue, ",") {
			switch tag {
			case "primary":
				f.Primary = true
			case "autoinc":
				col.options["autoinc"] = "AUTO_INCREMENT"
			case "null":
				f.Nullable = true
			default: // assume it's the name
				f.Name = tag
			}
		}
	}

	return nil
}

// reflectTypeToDefaultValue takes a go type and returns the default value for mysql.
func reflectTypeToDefaultValue(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int64, reflect.Bool:
		return "0"
	case reflect.String:
		return ""
	default:
		switch t.PkgPath() {
		case "time":
			switch t.Name() {
			case "Time":
				return "CURRENT_TIMESTAMP"
			default:
				fmt.Println("unknown struct from 'package time'")
				return ""
			}
		default:
			fmt.Println("[unknown package " + t.PkgPath() + "]")
			return ""
		}
	}
}

// reflectTypeToMySQLType converts a go type to a mysql type.
// There is some funky printing in here.
func reflectTypeToMySQLType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int64:
		return "INT"
	case reflect.Bool:
		return "TINYINT"
	case reflect.String:
		return "TEXT"
	default:
		switch t.PkgPath() {
		case "time":
			switch t.Name() {
			case "Time":
				return "TIMESTAMP"
			default:
				fmt.Println("unknown struct from 'package time'")
				return ""
			}
		default:
			fmt.Println("[unknown package " + t.PkgPath() + "]")
			return ""
		}
	}
}
