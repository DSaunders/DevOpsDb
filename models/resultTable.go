package models

import "golang.org/x/exp/slices"

// Should this be string/interface?
// Do we always cast the results to string.. probably not
type ResultTable = []map[string]string

func OnlyColumns(table ResultTable, columns []string) ResultTable {

	modifiedTable := table

	if len(columns) > 0 {
		for _, item := range modifiedTable {
			for column := range item {
				if !slices.Contains(columns, column) {
					delete(item, column)
				}
			}
		}
	}

	return modifiedTable
}
