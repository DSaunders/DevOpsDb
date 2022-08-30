package inputs

import (
	"devopsdb/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SqlTest struct {
	name   string
	query  string
	result models.Query
}

func TestSimpleSelects(t *testing.T) {

	tests := []SqlTest{
		{
			"slect all from one table",
			"select * from devops.builds",
			models.Query{
				SchemaName: "devops",
				Table:      "builds",
				Columns:    []string(nil),
				Limit:      0,
			},
		},
		{
			"select specific columns from one table",
			"select createdBy, started from devops.builds",
			models.Query{
				SchemaName: "devops",
				Table:      "builds",
				Columns:    []string{"createdby", "started"},
				Limit:      0,
			},
		},
		{
			"select all from one table with limit",
			"select * from devops.builds limit 10",
			models.Query{
				SchemaName: "devops",
				Table:      "builds",
				Columns:    []string(nil),
				Limit:      10,
			},
		},
	}

	for _, test := range tests {
		r, err := SqlToQuery(test.query)
		if err != nil {
			t.Errorf("Error parsing the query: %v", err)
		}
		assert.Equal(t, test.result, r, "Query '"+test.name+"' failed")
	}
}
