package djson

import "io"

type compileInstance struct {
	r          io.Reader
	tree       *tree
	bufferSize int
}

type Compiler struct {
	BufferSize int
}

func (compiler *Compiler) Compile(r io.Reader) error {
	ins := compileInstance{r: r, bufferSize: compiler.BufferSize, tree: newTree()}
	return ins.compile()
}

func (ins *compileInstance) compile() error {
	tokenGetter := NewTokenGetter(ins.r, uint(ins.bufferSize))
	token := &Token{}
	for {
		if err := tokenGetter.NextToken(token); err != nil {
			return err
		}
		if token.Type == TokenEOF {
			break
		}
		var err error
		switch ins.tree.curr.scope {
		case scopeGlobal:
			err = ins.handleGlobal(token)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (ins *compileInstance) handleGlobal(token *Token) error {
	switch token.Type {
	case TokenBlockStart:
		n := &node{}
		n.typ, n.scope = ins.of(token)
		ins.tree.curr = ins.tree.curr.add(n)
	case TokenBlockEnd:
		scope := ins.tree.curr.scope
		validEnds := []struct {
			scope nodeScope
			end   byte
		}{
			{scope: scopeMap, end: '}'},
			{scope: scopeTemplate, end: '>'},
			{scope: scopeNested, end: ')'},
			{scope: scopeArray, end: ']'},
		}
		for _, v := range validEnds {
			if scope == v.scope && token.Raw[0] == v.end {
				if ins.tree.curr.parent == nil {
					break
				}
				ins.tree.curr = ins.tree.curr.parent
				return nil
			}
		}
	case TokenVariable:
	}
	return ErrFromToken(UnexpectedChar, token)
}

func (ins *compileInstance) of(token *Token) (nodeType, nodeScope) {
	switch token.Raw[0] {
	case '{', '}':
		return nodeMap, scopeMap
	case '<', '>':
		return nodeTemplate, scopeTemplate
	case '(', ')':
		return nodeNested, scopeNested
	case '[', ']':
		return nodeArray, scopeArray
	}
	return nodeUnknown, scopeUnknown
}
