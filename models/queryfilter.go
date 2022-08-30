package models

import (
	"regexp"
	"strings"
)

type QueryFilter struct {
	Type      string // eq
	FieldName string // The name of the field to check
	Value     string // The value to compare against
	Children  []QueryFilter
}

func (f *QueryFilter) Filter(results ResultTable) ResultTable {

	var result = ResultTable{}

	switch f.Type {
	case "eq":
		var target = strings.ToLower(f.Value)
		for _, r := range results {
			if strings.ToLower(r[f.FieldName]) == target {
				result = append(result, r)
			}
		}
	case "ne":
		var target = strings.ToLower(f.Value)
		for _, r := range results {
			if strings.ToLower(r[f.FieldName]) != target {
				result = append(result, r)
			}
		}
	case "regex":
		regex := regexp.MustCompile("(?i)" + f.Value)
		for _, r := range results {
			if matched := regex.MatchString(r[f.FieldName]); matched {
				result = append(result, r)
			}
		}
	}

	return result
}
