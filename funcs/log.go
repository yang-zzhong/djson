package funcs

import "djson"

type log_ struct {
	*djson.CallableRegister
}

func NewLog() *log_ {
	l := &log_{CallableRegister: djson.NewCallableRegister("log")}
	l.RegisterCall("info", l.doInfo)
	l.RegisterCall("error", l.doError)
	l.RegisterCall("debug", l.doDebug)
	l.RegisterCall("fatal", l.doFatal)
	return l
}

func (h *log_) doInfo(val djson.Value, scanner djson.TokenScanner, vars djson.Context) (ret djson.Value, err error) {
	return
}
func (h *log_) doError(val djson.Value, scanner djson.TokenScanner, vars djson.Context) (ret djson.Value, err error) {
	return
}
func (h *log_) doDebug(val djson.Value, scanner djson.TokenScanner, vars djson.Context) (ret djson.Value, err error) {
	return
}
func (h *log_) doFatal(val djson.Value, scanner djson.TokenScanner, vars djson.Context) (ret djson.Value, err error) {
	return
}
