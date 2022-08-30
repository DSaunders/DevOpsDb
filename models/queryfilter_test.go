package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEquals(t *testing.T) {

	results := ResultTable{
		{"name": "bob", "age": "30"},
		{"name": "Bob", "age": "30"}, // test case-insensitivity
		{"name": "alice", "age": "40"},
		{"name": "herbert", "age": "19"},
	}

	results = (&QueryFilter{Type: "eq", FieldName: "name", Value: "bob"}).Filter(results)

	assert.Equal(t, 2, len(results))
	assert.Equal(t, "bob", results[0]["name"])
	assert.Equal(t, "Bob", results[1]["name"])
}

func TestNotEquals(t *testing.T) {

	results := ResultTable{
		{"name": "bob", "age": "30"},
		{"name": "Bob", "age": "30"}, // test case-insensitivity
		{"name": "alice", "age": "40"},
		{"name": "herbert", "age": "19"},
	}

	results = (&QueryFilter{Type: "ne", FieldName: "name", Value: "bob"}).Filter(results)

	assert.Equal(t, 2, len(results))
	assert.Equal(t, "alice", results[0]["name"])
	assert.Equal(t, "herbert", results[1]["name"])
}

func TestMatchesReges(t *testing.T) {

	results := ResultTable{
		{"name": "Peter", "age": "30"},
		{"name": "Bob Dole", "age": "30"},
		{"name": "saltpeter", "age": "19"},
		{"name": "sally field", "age": "40"},
	}

	results = (&QueryFilter{Type: "regex", FieldName: "name", Value: "^.*pete.*$"}).Filter(results)

	assert.Equal(t, 2, len(results))
	assert.Equal(t, "Peter", results[0]["name"])
	assert.Equal(t, "saltpeter", results[1]["name"])
}
