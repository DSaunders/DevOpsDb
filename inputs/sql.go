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

type queryVisitor struct {
	resultingQuery models.Query

	// Tracks whether we're in a binary expression (e.g. columnA='foo')
	// because we'll visit both sides separately
	inBinaryExpression bool
	binaryExpression   models.QueryFilter

	// This stack keeps track of the current nesting of and/ors
	binaryExpressionStack []models.QueryFilter
}

func (v *queryVisitor) Enter(in ast.Node) (ast.Node, bool) {

	switch node := in.(type) {

	case *ast.ColumnName:
		v.enterColumnNameNode(node)
	case *ast.TableName:
		v.enterTableNameNode(node)
	case *ast.Limit:
		v.enterLimitNode(node)
	case *ast.BinaryOperationExpr:
		v.enterBinaryExpressionNode(node)
	case *ast.PatternLikeExpr:
		v.enterLikeNode(node)
	case *ast.ValueExpr:
		v.enterValueNode(node)

	}

	return in, false
}

func (v *queryVisitor) enterColumnNameNode(node *ast.ColumnName) {
	if v.inBinaryExpression {
		// We're mid-where clause, so this is a column in an expression
		// (e.g. where x='foo')
		v.binaryExpression.FieldName = node.Name.L
		v.completeWhereClause()
	} else {
		// Otherwise, this must be a column name (e.g. select X from )
		v.resultingQuery.Columns = append(v.resultingQuery.Columns, node.Name.L)
	}
}

func (v *queryVisitor) enterTableNameNode(node *ast.TableName) {
	v.resultingQuery.SchemaName = node.Schema.L
	v.resultingQuery.Table = node.Name.L
}

func (v *queryVisitor) enterLimitNode(node *ast.Limit) {
	v.resultingQuery.Limit = int(node.Count.GetDatum().GetInt64())
}

func (v *queryVisitor) enterBinaryExpressionNode(node *ast.BinaryOperationExpr) {

	//  A normal node (e.g. x = 'foo' or x != 'foo')
	if node.Op == opcode.EQ || node.Op == opcode.NE {
		v.inBinaryExpression = true
		v.binaryExpression = models.QueryFilter{Type: node.Op.String()}
	}

	// A node that will nest other expressions (e.g 'A and B' or '(A or B) and C')
	if node.Op == opcode.LogicAnd || node.Op == opcode.LogicOr {

		opType := "and"
		if node.Op == opcode.LogicOr {
			opType = "or"
		}

		// Push a new node on to the stack
		v.binaryExpressionStack = append(
			v.binaryExpressionStack,
			models.QueryFilter{
				Type:     opType,
				Children: make([]models.QueryFilter, 0),
			},
		)
	}
}

func (v *queryVisitor) enterLikeNode(node *ast.PatternLikeExpr) {

	// We're entering a binary expression, but we'll get both sides
	// as individual visits to other nodes later
	v.inBinaryExpression = true
	v.binaryExpression = models.QueryFilter{Type: "regex"}
}

func (v *queryVisitor) enterValueNode(node *ast.ValueExpr) {

	// If we get a value and we're not in a expression
	// then I have no idea what to do!
	if !v.inBinaryExpression {
		return
	}

	v.binaryExpression.Value = node.GetDatum().GetString()

	// If we're in a 'like' then convert the wildcard string in to a regex
	// by replacing % with .* to make something like /^.*foo.*$/
	if v.binaryExpression.Type == "regex" {
		var regex = strings.Replace(v.binaryExpression.Value, "%", ".*", -1)
		v.binaryExpression.Value = "^" + regex + "$"
	}

	v.completeWhereClause()
}

func (v *queryVisitor) completeWhereClause() {

	// If either side of the expression is empty, we haven't seen both
	// nodes yet
	if v.binaryExpression.FieldName == "" || v.binaryExpression.Value == "" {
		return
	}

	// We're not in an AND/OR etc., so we can just add the condition to
	// the list of filters
	if len(v.binaryExpressionStack) == 0 {
		// otherwise, add us to the normal filters
		v.resultingQuery.Filters = append(v.resultingQuery.Filters, v.binaryExpression)
	}

	// We're in an AND/OR, so we need to add our expression to the item at the top of the stack
	// (identified as stack[length -1].. the last item in the slice)
	if len(v.binaryExpressionStack) > 0 {
		v.binaryExpressionStack[len(v.binaryExpressionStack)-1].Children = append(
			v.binaryExpressionStack[len(v.binaryExpressionStack)-1].Children,
			v.binaryExpression,
		)
	}

	v.inBinaryExpression = false
}

func (v *queryVisitor) Leave(in ast.Node) (ast.Node, bool) {

	binaryOp, isBinaryOperator := in.(*ast.BinaryOperationExpr)

	// We're not leaving a binary expression, so we don't care
	if !isBinaryOperator {
		return in, true
	}

	// The stack is empty, not sure what we're leaving here but there's
	// nothing to do
	if len(v.binaryExpressionStack) == 0 {
		return in, true
	}

	// We're leaving an expression we don't handle, so do nothing
	if binaryOp.Op != opcode.LogicAnd && binaryOp.Op != opcode.LogicOr {
		return in, true
	}

	// If we get here they we're leaving and AND/OR expression

	if len(v.binaryExpressionStack) >= 2 {
		// If we're *not* the last item in the stack, then this expression belongs as a child
		// of the one above us in the stack
		// e.g. the A or B is a child clause in the expression '(A or B) or C'
		v.binaryExpressionStack[len(v.binaryExpressionStack)-2].Children = append(
			v.binaryExpressionStack[len(v.binaryExpressionStack)-2].Children,
			v.binaryExpressionStack[len(v.binaryExpressionStack)-1],
		)
	} else {
		// This is the last item in the stack, so we're the outer expression
		// .. add our filter to the top-level filter list
		v.resultingQuery.Filters = append(v.resultingQuery.Filters, v.binaryExpressionStack[len(v.binaryExpressionStack)-1])
	}

	// Pop off the stack
	v.binaryExpressionStack = v.binaryExpressionStack[:len(v.binaryExpressionStack)-1]

	return in, true
}
