package connectors

import "devopsdb/models"

type Connector interface {
	GetSchemaForTable(table string) []string
	Get(query ConnectorQuery) models.ResultTable
}
