package inputs

import (
	"devopsdb/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleWhere(t *testing.T) {

	tests := []SqlTest{
		{
			"simple equals on one column",
			"select * from devops.builds where name = \"foo\"",
			models.Query{
				SchemaName: "devops",
				Table:      "builds",
				Columns:    []string(nil),
				Limit:      0,
				Filters: []models.QueryFilter{
					{Type: "eq", FieldName: "name", Value: "foo"},
				},
			},
		},
		{
			"simple equals on one column with reversed boolean logiv",
			"select * from devops.builds where 'foo' = name",
			models.Query{
				SchemaName: "devops",
				Table:      "builds",
				Columns:    []string(nil),
				Limit:      0,
				Filters: []models.QueryFilter{
					{Type: "eq", FieldName: "name", Value: "foo"},
				},
			},
		},
		{
			"simple equals on one column with column names specified",
			"select name, age from devops.builds where name = 'foo'",
			models.Query{
				SchemaName: "devops",
				Table:      "builds",
				Columns:    []string{"name", "age"},
				Limit:      0,
				Filters: []models.QueryFilter{
					{Type: "eq", FieldName: "name", Value: "foo"},
				},
			},
		},
		{
			"where like with wildcard end",
			"select name, age from devops.builds where name like 'foo%'",
			models.Query{
				SchemaName: "devops",
				Table:      "builds",
				Columns:    []string{"name", "age"},
				Limit:      0,
				Filters: []models.QueryFilter{
					{Type: "regex", FieldName: "name", Value: "^foo.*$"},
				},
			},
		},
		{
			"where like with wildcard start",
			"select name, age from devops.builds where name like '%foo'",
			models.Query{
				SchemaName: "devops",
				Table:      "builds",
				Columns:    []string{"name", "age"},
				Limit:      0,
				Filters: []models.QueryFilter{
					{Type: "regex", FieldName: "name", Value: "^.*foo$"},
				},
			},
		},
		{
			"where like with wildcard contains",
			"select name, age from devops.builds where name like '%foo%'",
			models.Query{
				SchemaName: "devops",
				Table:      "builds",
				Columns:    []string{"name", "age"},
				Limit:      0,
				Filters: []models.QueryFilter{
					{Type: "regex", FieldName: "name", Value: "^.*foo.*$"},
				},
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
