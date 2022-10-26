package djson

import "bytes"

type array []value

type arrayExecutor struct {
	getter    lexer
	variables *variables
	value     array
}

func newArrayExecutor(getter lexer, vs *variables) *arrayExecutor {
	return &arrayExecutor{
		getter:    getter,
		variables: vs,
	}
}

func (arr *array) set(idx int, val value) {
	(*arr)[idx] = val
}

func (arr *array) get(idx int) value {
	return (*arr)[idx]
}

func (arr *array) delAt(idx int) {
	*arr = append((*arr)[:idx], (*arr)[idx+1:]...)
}

func (arr *array) del(val ...value) {
	for i := 0; i < len(*arr); i++ {
		for _, v := range val {
			if (*arr)[i].equal(v) {
				arr.delAt(i)
				i--
			}
		}
	}
}

func (arr *array) append(val ...value) {
	*arr = append(*arr, val...)
}

func (arr *array) insertAt(idx int, val value) {
	tmp := (*arr)[idx:]
	*arr = append((*arr)[:idx], val)
	*arr = append(*arr, tmp...)
}

func (e *arrayExecutor) execute() (err error) {
	e.value, err = e.items()
	return
}

func (e *arrayExecutor) items() (val array, err error) {
	for {
		expr := newExpr(e.getter, [][]byte{{',', ']'}}, e.variables)
		if err = expr.execute(); err != nil {
			return
		}
		e.value.append(expr.value)
		if bytes.Equal(expr.endAt(), []byte{']'}) {
			return
		}
	}
}
