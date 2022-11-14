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
	p.scanner.PushEnds(TokenSemicolon)
	defer p.scanner.PopEnds(1)
	expr := NewStmt(p.scanner, vars)
	for {
		if err = expr.Execute(); err != nil {
			return
		}
		if expr.value.Type != ValueNull {
			val = expr.value
		}
		if expr.endAt() == TokenEOF {
			return
		}
	}
}
