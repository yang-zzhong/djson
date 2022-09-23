package djson

type ScopeType int
type VariableType int

const (
	ScopeGlobal = ScopeType(iota)
	ScopeObject
	ScopeArray
	ScopeTemplate
	ScopeExpr
	ScopeUnknown
)

const (
	VariableInt = VariableType(iota)
	VariableFloat
	VariableString
	VariableArray
	VariableObject
)

type Variable struct {
	Type  VariableType
	Value []byte
}

type Pair struct {
	Key   []byte
	Value Variable
}

type Scope struct {
	Type      ScopeType
	Variables map[string]Variable
	Pairs     []Pair
	Stash     []*Token
	Parent    *Scope
	Children  []*Scope
}

func (scope *Scope) Add(typ ScopeType) *Scope {
	child := &Scope{
		Type:      typ,
		Variables: make(map[string]Variable),
		Pairs:     []Pair{},
		Children:  []*Scope{},
		Parent:    scope,
	}
	scope.Children = append(scope.Children, scope)
	return child
}
