package sqlite

import (
	"github.com/chuckha/ttaal/storage"
)

type queryBuilder struct {
}

func (q *queryBuilder) Create(c storage.Creatable) string {
	return ""
}
