
## Usage
```golang
ctx := djson.NewContext()
ctx.Assign("_http", funcs.NewHttp())
ctx.Assign("_json", funcs.NewJson())

translator := djson.NewTranslator(Ctx(ctx))
data := `
_http.get({
    "url": "https://baidu.com",
    "body": _json.encode({"hello": "world"})
})
`
input := bytes.NewBuffer([]byte(data))
var output bytes.Buffer
if err := translator.Translate(input, &output); err != nil {
    panic(err)
}
```
