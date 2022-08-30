package models

type Query struct {
	SchemaName string
	Table      string
	Columns    []string
	Limit      int
	Filters    []QueryFilter
}

type QueryResult struct {
	Columns []string
	Results ResultTable
}
