package connectors

import "devopsdb/models"

type ConnectorQuery struct {
	TableName   string
	ColumnNames []string
	Top         int
	Filters     []models.QueryFilter
}
