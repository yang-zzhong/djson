package djson

type Parser interface {
	Parse() (val Value, vars *variables, err error)
}

type parser struct {
	scanner TokenScanner
}

func NewParser(ts TokenScanner) *parser {
	return &parser{scanner: ts}
}

func (p *parser) Parse() (val Value, vars *variables, err error) {
	vars = newVariables()
	for {
		expr := newStmt(p.scanner, vars)
		if err = expr.execute(); err != nil {
			return
		}
		val = expr.value
		if expr.endAt() == TokenEOF {
			return
		}
	}
}
