package djson

import (
	"bytes"
	"fmt"
	"net/http"
)

type httpc struct {
	*callableImp
}

func NewHttp() *httpc {
	h := &httpc{callableImp: newCallable("http")}
	h.register("get", h.getHttp)
	h.register("post", h.postHttp)
	h.register("put", h.putHttp)
	h.register("delete", h.deleteHttp)
	h.register("patch", h.patchHttp)
	h.register("head", h.headHttp)
	return h
}

func (h *httpc) getHttp(val Value, scanner TokenScanner, vars Context) (ret Value, err error) {
	return h.doHttp(val, scanner, vars, http.MethodGet)
}

func (h *httpc) putHttp(val Value, scanner TokenScanner, vars Context) (ret Value, err error) {
	return h.doHttp(val, scanner, vars, http.MethodPut)
}

func (h *httpc) postHttp(val Value, scanner TokenScanner, vars Context) (ret Value, err error) {
	return h.doHttp(val, scanner, vars, http.MethodPost)
}

func (h *httpc) deleteHttp(val Value, scanner TokenScanner, vars Context) (ret Value, err error) {
	return h.doHttp(val, scanner, vars, http.MethodDelete)
}

func (h *httpc) patchHttp(val Value, scanner TokenScanner, vars Context) (ret Value, err error) {
	return h.doHttp(val, scanner, vars, http.MethodPatch)
}

func (h *httpc) headHttp(val Value, scanner TokenScanner, vars Context) (ret Value, err error) {
	return h.doHttp(val, scanner, vars, http.MethodHead)
}

func (h *httpc) doHttp(val Value, scanner TokenScanner, vars Context, method string) (ret Value, err error) {
	stmt := NewStmtExecutor(scanner, vars)
	func() {
		scanner.PushEnds(TokenParenthesesClose)
		defer scanner.PopEnds(TokenParenthesesClose)
		err = stmt.Execute()
	}()
	if stmt.Exited() {
		Exit()
	}
	if err != nil {
		return
	}
	pv := stmt.Value().realValue()
	if pv.Type != ValueObject {
		err = fmt.Errorf("http only support a object as the params, current type is [%s]", val.TypeName())
		return
	}
	p := pv.Value.(Object)
	var url string
	if stringer, ok := p.Get([]byte{'u', 'r', 'l'}).Value.(Stringer); ok {
		url = stringer.String()
	} else {
		err = fmt.Errorf("only string can use as url, current type is [%s]", val.TypeName())
		return
	}
	var body []byte
	if byter, ok := p.Get([]byte{'b', 'o', 'd', 'y'}).Value.(Byter); ok {
		body = byter.Bytes()
	} else {
		err = fmt.Errorf("can't get a body from a [%s]", val.TypeName())
		return
	}
	var req *http.Request
	var res *http.Response
	if req, err = http.NewRequest(method, url, bytes.NewBuffer(body)); err != nil {
		return
	}
	cli := http.Client{}
	if res, err = cli.Do(req); err != nil {
		return
	}
	stmt = NewStmtExecutor(NewTokenScanner(NewLexer(res.Body, 128)), NewContext())
	stmt.scanner.PushEnds(TokenEOF)
	if err = stmt.Execute(); err == nil {
		ret = stmt.Value()
	}
	return
}
