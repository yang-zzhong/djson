package funcs

import (
	"bytes"
	"djson"
	"fmt"
	"io"
	"net/http"
)

type httpc struct {
	*djson.CallableRegister
}

func NewHttp() *httpc {
	h := &httpc{CallableRegister: djson.NewCallableRegister("http")}
	h.RegisterCall("get", h.getHttp)
	h.RegisterCall("post", h.postHttp)
	h.RegisterCall("put", h.putHttp)
	h.RegisterCall("delete", h.deleteHttp)
	h.RegisterCall("patch", h.patchHttp)
	h.RegisterCall("head", h.headHttp)
	return h
}

func (h *httpc) getHttp(val djson.Value, scanner djson.TokenScanner, vars djson.Context) (ret djson.Value, err error) {
	return h.doHttp(val, scanner, vars, http.MethodGet)
}

func (h *httpc) putHttp(val djson.Value, scanner djson.TokenScanner, vars djson.Context) (ret djson.Value, err error) {
	return h.doHttp(val, scanner, vars, http.MethodPut)
}

func (h *httpc) postHttp(val djson.Value, scanner djson.TokenScanner, vars djson.Context) (ret djson.Value, err error) {
	return h.doHttp(val, scanner, vars, http.MethodPost)
}

func (h *httpc) deleteHttp(val djson.Value, scanner djson.TokenScanner, vars djson.Context) (ret djson.Value, err error) {
	return h.doHttp(val, scanner, vars, http.MethodDelete)
}

func (h *httpc) patchHttp(val djson.Value, scanner djson.TokenScanner, vars djson.Context) (ret djson.Value, err error) {
	return h.doHttp(val, scanner, vars, http.MethodPatch)
}

func (h *httpc) headHttp(val djson.Value, scanner djson.TokenScanner, vars djson.Context) (ret djson.Value, err error) {
	return h.doHttp(val, scanner, vars, http.MethodHead)
}

func (h *httpc) doHttp(val djson.Value, scanner djson.TokenScanner, vars djson.Context, method string) (ret djson.Value, err error) {
	stmt := djson.NewStmtExecutor(scanner, vars)
	func() {
		scanner.PushEnds(djson.TokenParenthesesClose)
		defer scanner.PopEnds(djson.TokenParenthesesClose)
		err = stmt.Execute()
	}()
	if stmt.Exited() {
		djson.Exit()
	}
	if err != nil {
		return
	}
	pv := stmt.Value().RealValue()
	if pv.Type != djson.ValueObject {
		err = fmt.Errorf("http only support a object as the params, current type is [%s]", val.TypeName())
		return
	}
	p := pv.Value.(djson.Object)
	var req *http.Request
	var res *http.Response
	if req, err = h.request(p, method); err != nil {
		return
	}
	cli := http.Client{}
	if res, err = cli.Do(req); err != nil {
		return
	}
	return h.response(res)
}

func (h *httpc) response(res *http.Response) (ret djson.Value, err error) {
	header := djson.NewObject()
	for k, v := range res.Header {
		if len(v) == 1 {
			header.Set([]byte(k), djson.StringValue([]byte(v[0])...))
			continue
		}
		hv := make([]djson.Value, len(v))
		for i, vi := range v {
			hv[i] = djson.StringValue([]byte(vi)...)
		}
		header.Set([]byte(k), djson.ArrayValue(djson.NewArray(hv...)))
	}
	obj := djson.NewObject()
	obj.Set([]byte{'h', 'e', 'a', 'd', 'e', 'r'}, djson.ObjectValue(header))
	obj.Set([]byte{'s', 't', 'a', 't', 'u', 's', 'C', 'o', 'd', 'e'}, djson.IntValue(int64(res.StatusCode)))
	var body bytes.Buffer
	if _, err = io.Copy(&body, res.Body); err != nil {
		return
	}
	obj.Set([]byte{'b', 'o', 'd', 'y'}, djson.StringValue(body.Bytes()...))
	ret = djson.ObjectValue(obj)
	return
}

func (h *httpc) request(p djson.Object, method string) (req *http.Request, err error) {
	var url string
	if url, err = h.url(p); err != nil {
		return
	}
	if req, err = http.NewRequest(method, url, h.body(p)); err != nil {
		return
	}
	h.completeHeader(p, req)
	return req, nil
}

func (h *httpc) url(p djson.Object) (string, error) {
	url := p.Get([]byte{'u', 'r', 'l'}).RealValue()
	if stringer, ok := url.Value.(djson.Stringer); ok {
		return stringer.String(), nil
	}
	return "", fmt.Errorf("only string can use as url, current type is [%s]", url.TypeName())
}

func (h *httpc) body(p djson.Object) io.Reader {
	body := p.Get([]byte{'b', 'o', 'd', 'y'})
	if body.Type == djson.ValueNull {
		return nil
	}
	if byter, ok := body.Value.(djson.Byter); ok {
		return bytes.NewBuffer(byter.Bytes())
	}
	return nil
}

func (h *httpc) completeHeader(p djson.Object, req *http.Request) {
	header := p.Get([]byte{'h', 'e', 'a', 'd', 'e', 'r'}).RealValue()
	if header.Type != djson.ValueObject {
		return
	}
	header.Value.(djson.Object).Each(func(k []byte, val djson.Value) bool {
		if val.Type == djson.ValueArray {
			val.Value.(djson.Array).Each(func(i int, val djson.Value) bool {
				if stringer, ok := val.Value.(djson.Stringer); ok {
					req.Header.Add(string(k), stringer.String())
				}
				return true
			})
			return true
		}
		if stringer, ok := val.Value.(djson.Stringer); ok {
			req.Header.Set(string(k), stringer.String())
		}
		return true
	})
}
