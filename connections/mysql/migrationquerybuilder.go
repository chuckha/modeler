package mysql

import (
	"fmt"
	"reflect"
	"strings"
)

type metaQueryBuilder struct {
	logger
}

const (
	existsQueryFormat  = "select 1 from tables where table_schema = '%s' and table_name = '%s' limit 1;"
	columnsQueryFormat = "select COLUMN_NAME, COLUMN_DEFAULT, IS_NULLABLE, DATA_TYPE from columns where table_schema='%s' and table_name='%s';"
	// CURRENT PAIN: can't join constraints to column table ?? :( or need multiple queries...
	keysQueryFormat = "select constraint_name, column_name from key_column_usage where table_schema='%s' and table_name='%s';"
)

// TableExistsQuery is the query to determine if a table is already created.
func (m *metaQueryBuilder) TableExistsQuery(db, table string) string {
	q := fmt.Sprintf(existsQueryFormat, db, table)
	m.logger.Debugln("Exists query:", q)
	return q
}

// CreateTableQuery returns the create table query for the model type.
func (m *metaQueryBuilder) CreateTableQuery(tableName string, model interface{}) string {
	tab := &table{
		name: tableName,
		cols: make([]*col, 0),
	}
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
		col := newCol(sf.Type)
		for _, tag := range strings.Split(sqlTagValue, ",") {
			col.Add(tag)
		}
		tab.cols = append(tab.cols, col)
	}
	createQuery := tab.createQuery()
	m.logger.Debugln(createQuery)
	return createQuery
}

// TableDataQuery returns a query that will get the data representation of a table.
func (m *metaQueryBuilder) TableDataQuery(database, table string) string {
	return fmt.Sprintf(columnsQueryFormat, database, table)
}

// table is the metadata information of a table.
// Things that relate to no particular field go here.
// Possible additions include engine, constraints, and encoding.
type table struct {
	name string
	cols []*col
}

// createQuery will generate a create table statement.
// All necessary information to build the create table statement is available to this function.
func (t *table) createQuery() string {
	primaryKeys := make([]string, 0)
	createLines := make([]string, 0)
	out := fmt.Sprintf("CREATE TABLE %s.%s (\n", databaseName, t.name)
	for _, col := range t.cols {
		if col.primary {
			primaryKeys = append(primaryKeys, fmt.Sprintf("`%v`", col.name))
		}
		createLines = append(createLines, col.String())
	}
	if len(primaryKeys) > 0 {
		createLines = append(createLines, fmt.Sprintf("PRIMARY KEY(%v)", strings.Join(primaryKeys, ",")))
	}
	return out + strings.Join(createLines, ",\n") + "\n)"
}

// col is one column in a table.
// It is kind of like the glue between go struct field and mysql field.
type col struct {
	name         string
	mysqlType    string
	defaultValue string
	primary      bool
	options      map[string]string
}

// newCol creates and configures a col struct from a reflect.Type.
// Anything that is a function of the type of field can go in here. For example, default value.
func newCol(sf reflect.Type) *col {
	return &col{
		// start with defaults, override in add
		mysqlType: reflectTypeToMySQLType(sf),
		options: map[string]string{
			"default": reflectTypeToDefaultValue(sf),
		},
	}
}

// Add configures a col based on the struct tag passed in.
func (c *col) Add(tag string) {
	switch tag {
	case "primary":
		c.primary = true
	case "autoinc":
		c.options["autoinc"] = "AUTO_INCREMENT"
	case "null":
		c.options["nullable"] = "NULL"
	default: // assume it's the name
		c.name = tag
	}
}

// String turns a configured col into a single line to be embedded in a create table query.
func (c *col) String() string {
	out := make([]string, 0)
	out = append(out, fmt.Sprintf("`%v`", c.name))
	out = append(out, c.mysqlType)
	//  Delete the default if this is a primary key
	if c.primary {
		delete(c.options, "default")
		delete(c.options, "nullable")
	}
	for k, v := range c.options {
		if v == "" {
			continue
		}
		if k == "default" {
			out = append(out, fmt.Sprintf("DEFAULT %s", v))
			continue
		}
		out = append(out, v)
	}
	return strings.Join(out, " ")
}
