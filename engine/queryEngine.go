package engine

import (
	"devopsdb/connectors"
	"devopsdb/models"
	"devopsdb/utils"
)

func New() *QueryEngine {
	return &QueryEngine{
		connectors: make(map[string]connectors.Connector, 0),
	}
}

type QueryEngine struct {
	connectors map[string]connectors.Connector
}

func (engine *QueryEngine) AddConnector(schemaName string, conn connectors.Connector) {
	// This should check for duplicates.. just in case
	engine.connectors[schemaName] = conn
}

func (engine *QueryEngine) Execute(query models.Query) *models.QueryResult {
	// TODO: we need to combine any joins

	connector := engine.connectors[query.SchemaName]
	// TODO: what if it isn't found?

	results := connector.Get(connectors.ConnectorQuery{
		TableName:   query.Table,
		ColumnNames: query.Columns,
		Filters:     query.Filters, // TODO: only the ones that relate to this table
	})

	if query.Limit != 0 {
		resultsToReturn := utils.Min(query.Limit, len(results))
		results = results[:resultsToReturn]
	}

	// We set the columns here, becuase when there are multiple providers
	// involved we'll be the only place that knows the full list of columns
	// returned (plus we can remove aliases etc.)
	returnedColumns := query.Columns
	if len(query.Columns) == 0 {
		returnedColumns = connector.GetSchemaForTable(query.Table)
	}

	return &models.QueryResult{
		Columns: returnedColumns,
		Results: results,
	}
}
