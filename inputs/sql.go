package inputs

import (
	"devopsdb/models"
	"log"
	"strings"

	"github.com/blastrain/vitess-sqlparser/tidbparser/ast"
	"github.com/blastrain/vitess-sqlparser/tidbparser/parser"
	"github.com/blastrain/vitess-sqlparser/tidbparser/parser/opcode"
)

func SqlToQuery(query string) (models.Query, error) {
	p := parser.New()

	stmtNodes, err := p.Parse(query, "", "")
	if err != nil {
		log.Fatal(err)
	}

	visitor := &queryVisitor{}
	stmtNodes[0].Accept(visitor)

	return visitor.resultingQuery, nil
}

// This is the thing that visits each node and populates the query
type queryVisitor struct {
	resultingQuery  models.Query
	isInWhereClause bool
	whereClause     models.QueryFilter
	logicStack      []models.QueryFilter
}

func (v *queryVisitor) Enter(in ast.Node) (ast.Node, bool) {
	if columnName, ok := in.(*ast.ColumnName); ok {
		// If we're mid-where clause
		if v.isInWhereClause {
			v.whereClause.FieldName = columnName.Name.L

			v.completeWhereClause()
		} else {
			// Otherwise, this must be a column name
			v.resultingQuery.Columns = append(v.resultingQuery.Columns, columnName.Name.L)
		}
	}

	if table, ok := in.(*ast.TableName); ok {
		v.resultingQuery.SchemaName = table.Schema.L
		v.resultingQuery.Table = table.Name.L
	}

	if limit, ok := in.(*ast.Limit); ok {
		v.resultingQuery.Limit = int(limit.Count.GetDatum().GetInt64())
	}

	if where, ok := in.(*ast.BinaryOperationExpr); ok {
		if where.Op == opcode.LogicAnd {
			// Push a new AND on to the stack
			v.logicStack = append(
				v.logicStack,
				models.QueryFilter{Type: "and", Children: make([]models.QueryFilter, 0)},
			)
		} else if where.Op == opcode.LogicOr {
			// Push a new AND on to the stack
			v.logicStack = append(
				v.logicStack,
				models.QueryFilter{Type: "or", Children: make([]models.QueryFilter, 0)},
			)
		} else {
			v.isInWhereClause = true
			v.whereClause = models.QueryFilter{Type: where.Op.String()}
		}
	}

	if _, ok := in.(*ast.PatternLikeExpr); ok {
		v.isInWhereClause = true
		v.whereClause = models.QueryFilter{Type: "regex"}

		v.completeWhereClause()
	}

	if value, ok := in.(*ast.ValueExpr); ok {
		if v.isInWhereClause {
			v.whereClause.Value = value.GetDatum().GetString()

			if v.whereClause.Type == "regex" {
				// This is a 'like', convert it to a regex
				var regex = strings.Replace(v.whereClause.Value, "%", ".*", -1)
				v.whereClause.Value = "^" + regex + "$"
			}

			v.completeWhereClause()
		}
	}

	return in, false
}

func (v *queryVisitor) completeWhereClause() {
	if v.whereClause.FieldName != "" && v.whereClause.Value != "" {
		if len(v.logicStack) > 0 {
			// If we're in an AND statement, add ourselves to the parent
			v.logicStack[len(v.logicStack)-1].Children = append(v.logicStack[len(v.logicStack)-1].Children, v.whereClause)
		} else {
			// otherwise, add us to the normal filters
			v.resultingQuery.Filters = append(v.resultingQuery.Filters, v.whereClause)
		}
		v.isInWhereClause = false
	}
}

func (v *queryVisitor) Leave(in ast.Node) (ast.Node, bool) {

	if boolOperator, ok := in.(*ast.BinaryOperationExpr); ok && len(v.logicStack) > 0 {

		if boolOperator.Op == opcode.LogicAnd || boolOperator.Op == opcode.LogicOr {

			if len(v.logicStack) >= 2 {
				// Is there one above me? If so append me to their children
				v.logicStack[len(v.logicStack)-2].Children = append(
					v.logicStack[len(v.logicStack)-2].Children,
					v.logicStack[len(v.logicStack)-1],
				)
			} else {
				// I'm the last one, add me to the overall filters
				v.resultingQuery.Filters = append(v.resultingQuery.Filters, v.logicStack[len(v.logicStack)-1])
			}

			// Pop off the stack
			v.logicStack = v.logicStack[:len(v.logicStack)-1]
		}
	}

	return in, true
}
