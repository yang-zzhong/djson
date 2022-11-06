package djson

type str struct {
	bytes []byte
	*callableImp
}

func newString(bs ...byte) *str {
	s := &str{bytes: bs, callableImp: newCallable("string")}
	s.register("index", indexString)
	s.register("sub", indexString)
	return s
}

func indexString(val value, scanner *tokenScanner, vars *variables) (ret value, err error) {
	return
}

func subString(val value, scanner *tokenScanner, vars *variables) (ret value, err error) {
	return
}
