package engine

import (
	"devopsdb/connectors"
	"devopsdb/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

// e.g. "select * from azureDevOps.builds"
func TestSelectAllFromSingleEntityReturnsEverything(t *testing.T) {

	engine, _ := createEngine()

	// Act
	result := engine.Execute(
		models.Query{SchemaName: "azureDevOps", Table: "builds"},
	)

	// Assert
	assert.Equal(t, 2, len(result.Results))
	assert.Equal(t, "bob", result.Results[0]["startedby"])
	assert.Equal(t, "alice", result.Results[1]["startedby"])
}

// e.g. "select started, ended from azureDevOps.builds"
func TestSelectSpecificColumnsFromSingleEntity(t *testing.T) {

	engine, _ := createEngine()

	// Act
	result := engine.Execute(
		models.Query{
			SchemaName: "azureDevOps",
			Table:      "builds",
			Columns:    []string{"started", "ended"}},
	)

	// Assert
	for _, result := range result.Results {
		if result["started"] == "" || result["ended"] == "" || result["startedby"] != "" {
			t.Fatal("The column filters were not passed to the connector")
		}
	}
}

// e.g. "select * from azureDevOps.builds limit 10"
func TestSelectWithLimitFromSingleEntity(t *testing.T) {

	engine, _ := createEngine()

	// Act
	result := engine.Execute(
		models.Query{
			SchemaName: "azureDevOps",
			Table:      "builds",
			Limit:      1,
		},
	)

	assert.Equal(t, 1, len(result.Results))
	assert.Equal(t, "bob", result.Results[0]["startedby"])
}

func TestReturnsAllColumnsWithResultsWhenSelectAll(t *testing.T) {
	engine, _ := createEngine()

	result := engine.Execute(
		models.Query{
			SchemaName: "azureDevOps",
			Table:      "builds",
		},
	)

	assert.Equal(t, []string{"startedby", "started", "ended"}, result.Columns)
}

func TestReturnsSelectedColumnsInOrderWithResults(t *testing.T) {

	engine, _ := createEngine()

	result := engine.Execute(
		models.Query{
			SchemaName: "azureDevOps",
			Table:      "builds",
			Columns:    []string{"ended", "started"},
		},
	)

	assert.Equal(t, []string{"ended", "started"}, result.Columns)
}

func TestPassesWhereClauseToConnector(t *testing.T) {

	engine, connector := createEngine()

	engine.Execute(
		models.Query{
			SchemaName: "azureDevOps",
			Table:      "builds",
			Columns:    []string{"startedby"},
			Filters: []models.QueryFilter{
				{FieldName: "startedby", Type: "eq", Value: "bob"},
			},
		},
	)

	assert.Equal(t, []models.QueryFilter{{FieldName: "startedby", Type: "eq", Value: "bob"}}, connector.PassedQueryFilters)
}

func createEngine() (*QueryEngine, *FakeConnector) {
	engine := New()
	conn := &FakeConnector{}
	engine.AddConnector("azureDevOps", conn)
	return engine, conn
}

type FakeConnector struct {
	PassedQueryFilters []models.QueryFilter
}

func (f *FakeConnector) GetSchemaForTable(table string) []string {
	return []string{"startedby", "started", "ended"}
}

func (f *FakeConnector) Get(query connectors.ConnectorQuery) models.ResultTable {

	f.PassedQueryFilters = query.Filters

	r := models.ResultTable{
		0: map[string]string{
			"startedby": "bob",
			"started":   "monday",
			"ended":     "wednesday",
		},
		1: map[string]string{
			"startedby": "alice",
			"started":   "thursday",
			"ended":     "friday",
		},
	}

	if len(query.ColumnNames) > 0 {
		for _, item := range r {
			for column := range item {
				if !slices.Contains(query.ColumnNames, column) {
					delete(item, column)
				}
			}
		}
	}

	return r
}
