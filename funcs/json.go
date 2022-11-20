package funcs

import (
	"bytes"
	"djson"
	"fmt"
)

type jsonc struct {
	*djson.CallableRegister
}

func NewJson() *jsonc {
	j := &jsonc{CallableRegister: djson.NewCallableRegister("json")}
	j.RegisterCall("decode", j.decode)
	j.RegisterCall("encode", j.encode)
	return j
}

func (h *jsonc) decode(val djson.Value, scanner djson.TokenScanner, vars djson.Context) (ret djson.Value, err error) {
	if byter, ok := val.Value.(djson.Byter); ok {
		scanner := djson.NewTokenScanner(djson.NewLexer(bytes.NewBuffer(byter.Bytes()), 128))
		stmt := djson.NewStmtExecutor(scanner, vars)
		if err = stmt.Execute(); err != nil {
			return
		}
		if stmt.Exited() {
			djson.Exit()
		}
		ret = stmt.Value()
		return
	}
	err = fmt.Errorf("decode json only support a djson string")
	return
}

func (h *jsonc) encode(val djson.Value, scanner djson.TokenScanner, vars djson.Context) (ret djson.Value, err error) {
	encoder := djson.NewJsonEncoder()
	var buf bytes.Buffer
	if _, err = encoder.Encode(val, &buf); err != nil {
		return
	}
	ret = djson.StringValue(buf.Bytes()...)
	return
}
