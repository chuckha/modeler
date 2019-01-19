package storage

import (
	"database/sql"
	"fmt"
)

// Creatable is the set of behaviors for generating a Create statement.
// The models used for your application must implement this interface.
type Creatable interface {
	Table() string
	Fields() []string
	Values() []interface{}
}

// QueryBuilder describes the types of queries the implementor needs to implement.
// Each implementation is specific to the sql backend.
type QueryBuilder interface {
	Create(Creatable) string
}

// Database is the direct connection to a database backend.
type Database interface {
	Exec(string, ...interface{}) (sql.Result, error)
}

// Storage ties together the connection and the query builder.
type Storage struct {
	Database
	QueryBuilder
}

// New returns a configured Storage.
func New(db Database, qb QueryBuilder) *Storage {
	return &Storage{
		Database:     db,
		QueryBuilder: qb,
	}
}

// Create saves the creatable and returns the id of the newly created object.
// Create does assumes creatable does not exist.
func (s *Storage) Create(creatable Creatable) (int64, error) {
	query := s.QueryBuilder.Create(creatable)
	r, err := s.Exec(query, creatable.Values()...)
	if err != nil {
		return int64(-1), fmt.Errorf("failed to run query %q: %v", query, err)
	}
	return r.LastInsertId()
}
