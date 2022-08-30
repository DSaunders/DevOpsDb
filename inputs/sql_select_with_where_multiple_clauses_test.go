package inputs

import (
	"devopsdb/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhereWithAnd(t *testing.T) {

	tests := []SqlTest{
		{
			"select with simple AND equals",
			"select * from devops.builds where name='foo' and age='23'",
			models.Query{
				SchemaName: "devops",
				Table:      "builds",
				Columns:    []string(nil),
				Limit:      0,
				Filters: []models.QueryFilter{
					{
						Type: "and",
						Children: []models.QueryFilter{
							{Type: "eq", FieldName: "name", Value: "foo"},
							{Type: "eq", FieldName: "age", Value: "23"},
						},
					},
				},
			},
		},
		{
			"select with nested AND equals",
			"select * from devops.builds where name='foo' and age='23' and colour='red'",
			models.Query{
				SchemaName: "devops",
				Table:      "builds",
				Columns:    []string(nil),
				Limit:      0,
				Filters: []models.QueryFilter{
					{
						Type: "and",
						Children: []models.QueryFilter{
							{
								Type: "and",
								Children: []models.QueryFilter{
									{Type: "eq", FieldName: "name", Value: "foo"},
									{Type: "eq", FieldName: "age", Value: "23"},
								},
							},
							{Type: "eq", FieldName: "colour", Value: "red"},
						},
					},
				},
			},
		},
		{
			"select with nested AND and OR equals",
			"select * from devops.builds where (name='foo' and age='23') or (name='bar' and age='40')",
			models.Query{
				SchemaName: "devops",
				Table:      "builds",
				Columns:    []string(nil),
				Limit:      0,
				Filters: []models.QueryFilter{
					{
						Type: "or",
						Children: []models.QueryFilter{
							{
								Type: "and",
								Children: []models.QueryFilter{
									{Type: "eq", FieldName: "name", Value: "foo"},
									{Type: "eq", FieldName: "age", Value: "23"},
								},
							},
							{
								Type: "and",
								Children: []models.QueryFilter{
									{Type: "eq", FieldName: "name", Value: "bar"},
									{Type: "eq", FieldName: "age", Value: "40"},
								},
							},
						},
					},
				},
			},
		},
		{
			"select with multiple ANDs and no brackets",
			"select * from devops.builds where name='foo' and age='23' and colour='red'",
			models.Query{
				SchemaName: "devops",
				Table:      "builds",
				Columns:    []string(nil),
				Limit:      0,
				Filters: []models.QueryFilter{
					{
						Type: "and",
						Children: []models.QueryFilter{
							{
								Type: "and",
								Children: []models.QueryFilter{
									{Type: "eq", FieldName: "name", Value: "foo"},
									{Type: "eq", FieldName: "age", Value: "23"},
								},
							},
							{Type: "eq", FieldName: "colour", Value: "red"},
						},
					},
				},
			},
		},
		{
			"select with ANDs with a like",
			"select * from devops.builds where name='foo' and colour like 'red%'",
			models.Query{
				SchemaName: "devops",
				Table:      "builds",
				Columns:    []string(nil),
				Limit:      0,
				Filters: []models.QueryFilter{
					{
						Type: "and",
						Children: []models.QueryFilter{
							{Type: "eq", FieldName: "name", Value: "foo"},
							{Type: "regex", FieldName: "colour", Value: "^red.*$"},
						},
					},
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

// TODO: check devops connector, how will it find the project setting? Must be in a top-level filter or the result of a top-level AND?
