package models

import (
	"regexp"
	"strings"
)

type QueryFilter struct {
	Type      string // eq ('equal'), ne ('not equal'), 'regex', 'and', 'or'
	FieldName string // The name of the field to check
	Value     string // The value to compare against
	Children  []QueryFilter // Inner conditions for and/or nodes
}

func (f *QueryFilter) Filter(results ResultTable) ResultTable {

	var result = ResultTable{}

	for _, row := range results {
		if f.rowPasses(row) {
			result = append(result, row)
		}
	}

	return result
}

func (f *QueryFilter) rowPasses(row map[string]string) bool {
	switch f.Type {

	case "eq":
		var target = strings.ToLower(f.Value)
		return strings.ToLower(row[f.FieldName]) == target

	case "ne":
		var target = strings.ToLower(f.Value)
		return strings.ToLower(row[f.FieldName]) != target

	case "regex":
		regex := regexp.MustCompile("(?i)" + f.Value)
		return regex.MatchString(row[f.FieldName])

	case "and":
		passes := true
		for _, child := range f.Children {
			if !child.rowPasses(row) {
				passes = false
			}
		}
		return passes

	case "or":
		for _, child := range f.Children {
			if child.rowPasses(row) {
				return true
			}
		}
		return false
	}

	return false
}
