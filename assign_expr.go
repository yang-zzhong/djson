package djson

type assignExpr struct {
	ref   []byte
	value expr
	scope *scope
}

func (a *assignExpr) Eval() (interface{}, error) {

}
