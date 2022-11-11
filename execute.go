package djson

type executor struct {
	scanner *tokenScanner
	vars    variables
}

func (e *executor) execute() (val value, err error) {
	for {
		expr := newStmt(e.scanner, &e.vars)
		if err = expr.execute(); err != nil {
			return
		}
		val = expr.value
		if expr.endAt() == TokenEOF {
			return
		}
	}
}
