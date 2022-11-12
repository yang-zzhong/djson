## What is djson

dynamic json, it's a self-explain json, just like bellow

```
user = {
    "name": "Oliver",
    "desc": "emmmm, nothing to show around",
    "age": 106
}

{
    "username": user.name,
    "age": user.age + 10,
    "desc": user.desc + " PLUS: we just modify the age a little bit"
}
```

it's not just like that, we provide more feature to djson, such as native function call, see the example bellow

```
# we declare a variable, then we can change it to control the djson behavior

opt = ""

ifjjname = opt == "JJ" => "JJ's greate name"

users = [
    "JJ",
    "Mike",
    "Miky"
]

users.filter(ifjjname != null && v == "JJ")

```

## basic usage

```golang
translator := NewTranslator(NewJsonEncoder("  "), BuffSize(1024))
input := []byte("\"hello\"")
ib := bytes.NewBuffer(input)
ob := bytes.Buffer{}
_, err := translator.Translate(ib, &ob)
if err != nil {
	t.Fatal(err)
}
```
