package djson

type expr interface {
	Eval() interface{}
}

type program struct {
	scope *scope
	exprs []expr
}

func (p *program) Execute() error {

}

func (p *program) appendExpr(e expr) {
	p.exprs = append(p.exprs, e)
}
