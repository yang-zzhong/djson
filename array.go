package djson

type array []value

func (arr *array) set(idx int, val value) {
	(*arr)[idx] = val
}

func (arr *array) get(idx int) value {
	return (*arr)[idx]
}

func (arr *array) del(idx int) {
	*arr = append((*arr)[:idx], (*arr)[idx+1:]...)
}

func (arr *array) append(val value) {
	*arr = append(*arr, val)
}

func (arr *array) insertAt(idx int, val value) {
	tmp := (*arr)[idx:]
	*arr = append((*arr)[:idx], val)
	*arr = append(*arr, tmp...)
}
