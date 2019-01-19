package sqlite

import (
	"github.com/chuckha/modeler/storage"
)

type queryBuilder struct {
}

func (q *queryBuilder) Create(c storage.Creatable) string {
	return ""
}
