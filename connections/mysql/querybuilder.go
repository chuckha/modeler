package mysql

import (
	"fmt"
	"strings"

	"github.com/chuckha/modeler/storage"
)

type queryBuilder struct {
	log logger
}

// Create returns the insert command with ? in place of values.
// values will be substituted at a later time.
func (q *queryBuilder) Create(creatable storage.Creatable) string {
	query := fmt.Sprintf(createQuery, creatable.Table(), strings.Join(creatable.Fields(), ", "), getPlaceholders(len(creatable.Values())))
	q.log.Debugf("Create query: %q\n", query)
	return query
}

// returns the number of necessary placeholders separated by commas
func getPlaceholders(num int) string {
	placeholders := make([]string, num)
	for i := range placeholders {
		placeholders[i] = queryPlaceholder
	}
	return strings.Join(placeholders, ",")
}
