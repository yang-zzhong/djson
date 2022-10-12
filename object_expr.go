package djson

type pairExpr struct {
	key   expr
	val   expr
	scope *scope
}

type objectExpr struct {
	pairs []pairExpr
}
