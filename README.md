## What is djson

dynamic json, it's a self-explain json, just like bellow

```
user = {
  "name": "Oliver",
  "desc": "emmmm, nothing to show around",
  "age": 106
};

{
  "username": user.name,
  "age": user.age + 10,
  "desc": user.desc + " PLUS: we just modify the age a little bit"
};

```

it's not just like that, we provide more feature to djson, such as native function call, see the example bellow

```
# we declare a variable, then we can change it to control the djson behavior

opt = "";

jjname = opt == "JJ" => "JJ's greate name";

users = [
  "JJ",
  "Mike",
  "Miky"
];

users.filter(jjname != null && v == "JJ");

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

## grammar

assignation
```
a = "hello";

b = ["item1", "item2"];

c = {
  "key": "value"
};

d = true;

e = false;

f = a;

```

reduction expr
```
# bool => expr
# such as

a = "hello" == var => "var is hello"; # if var not equal "hello", a will be null
```

string native funcs

```

idx = str.index(sub);

str = str1 + str2;

str = str1 - str2; # replace all str2 in str1 with ""

```

array native funcs
```
newArr = arr.filter(i == 0 || v == "hello");

arr.0 = "world";

item = arr.0;

newArr = arr.map({
  "idx": i,
  "item": v
});

arr.set(i == 0 || v == "hello" => "new hello");

arr.del(i == 0 || v == "hello");

arr1 - arr2 # delete items match arr2 in arr1;

arr1 + arr2;

```
object native funcs

```

newObj = obj.filter(k == "0" || v == "hello");

newObj.0 = "world";

val = newObj.0;

obj.set(k == "0" || v == "hello" => "new hello");

obj.del(k == "0" || v == "hello");
obj1 - obj2 # delete items match arr2 in arr1;

obj1 + obj2;

```

