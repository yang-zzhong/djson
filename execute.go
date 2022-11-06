package djson

type executor struct {
	scanner *tokenScanner
	vars    variables
}

func (e *executor) execute() (val value, err error) {
	e.scanner.pushEnds(TokenEOF)
	defer e.scanner.popEnds(1)
	for {
		expr := newExpr(e.scanner, &e.vars)
		if err = expr.execute(); err != nil {
			return
		}
		val = expr.value
	}
	return
}
