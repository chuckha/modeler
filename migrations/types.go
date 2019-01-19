package migrations

import (
	"database/sql"

	"github.com/chuckha/modeler/repr"
	"github.com/chuckha/modeler/storage"
)

// MigratorDatabase extends the typical storage.Database interface.
// Another way to think about it is that it's the same as a storage.Database but with an extended set of functions.
// I expect storage.Database will have all of these methods and this interface can collapse.
type MigratorDatabase interface {
	storage.Database
	QueryRow(string, ...interface{}) *sql.Row
	Query(string, ...interface{}) (*sql.Rows, error)
}

// Representer can convert sql data and go interfaces to the same type.
// This makes the tables comparable.
type Representer interface {
	RepresentationFromRows(*sql.Rows) (*repr.Table, error)
	RepresentationFromInterface(interface{}) *repr.Table
}

// MetaQueryBuilder describes the behavior required for generating metaqueries (questions about the database/tables themselves).
type MetaQueryBuilder interface {
	TableExistsQuery(string, string) string
	CreateTableQuery(string, interface{}) string
	TableDataQuery(string, string) string
}

// Migrator ties together the connection and the query builder.
type Migrator struct {
	Database         MigratorDatabase
	MetaQueryBuilder MetaQueryBuilder
	Representer      Representer
}

// New returns a migrator
func New(db MigratorDatabase, mqb MetaQueryBuilder) *Migrator {
	return &Migrator{
		Database:         db,
		MetaQueryBuilder: mqb,
	}
}

// TableExists returns true if the table exists and false if it does not.
func (m *Migrator) TableExists(db, tableName string) bool {
	query := m.MetaQueryBuilder.TableExistsQuery(db, tableName)
	row := m.Database.QueryRow(query)
	var out int
	// Note: Hiding errors will eventually cause issues.
	return row.Scan(&out) == nil
}

// CreateTable generates a create table query and executes it.
func (m *Migrator) CreateTable(tableName string, instance interface{}) error {
	query := m.MetaQueryBuilder.CreateTableQuery(tableName, instance)
	_, err := m.Database.Exec(query)
	return err
}

// TableUpToDate returns true if the existing table is the same as the model.
// This assumes the model is the most up-to-date representation desired.
func (m *Migrator) TableUpToDate(db, tableName string, model interface{}) bool {
	// query information schema to get a sql.result
	query := m.MetaQueryBuilder.TableDataQuery(db, tableName)
	rows, err := m.Database.Query(query)
	// TODO fix the error handling here.
	if err != nil {
		return false
	}
	existing, err := m.Representer.RepresentationFromRows(rows)
	if err != nil {
		return false
	}
	current := m.Representer.RepresentationFromInterface(model)
	return existing.Same(current)
}
