package djson

import "bytes"

type Context interface {
	Assign(varName []byte, val Value)
	ValueOf(name []byte) Value
	PushScope()
	PopScope()
	Merge(ctx Context)
	All() []Variable
	Copy() Context
	pushMe(val Value)
	popMe()
}

type Variable struct {
	Name  []byte
	Value Value
}

type scope struct {
	p    *scope
	vars []Variable
}

type ctx struct {
	scope *scope
}

var _ Context = &ctx{}

func NewContext(vars ...Variable) *ctx {
	s := &scope{vars: vars}
	return &ctx{
		scope: s,
	}
}

func (s *scope) assign(name []byte, val Value) {
	if idx := s.indexOf(name); idx > -1 {
		s.vars[idx] = Variable{Name: name, Value: val}
		return
	}
	s.vars = append(s.vars, Variable{Name: name, Value: val})
}

func (s *scope) copy() *scope {
	var p *scope
	if s.p != nil {
		p = s.p.copy()
	}
	r := &scope{
		p:    p,
		vars: make([]Variable, len(s.vars)),
	}
	copy(r.vars, s.vars)
	return s
}

func (s *scope) indexOf(name []byte) int {
	for i := range s.vars {
		if bytes.Equal(s.vars[i].Name, name) {
			return i
		}
	}
	return -1
}

func (s *scope) del(idx int) {
	s.vars = append(s.vars[:idx], s.vars[idx+1:]...)
}

func (v *ctx) Assign(name []byte, val Value) {
	scope := v.scope
	for scope != nil {
		if idx := scope.indexOf(name); idx > -1 {
			scope.vars[idx] = Variable{Name: name, Value: val}
			return
		}
		scope = scope.p
	}
	v.scope.assign(name, val)
}

func (ctx *ctx) Merge(lv Context) {
	all := lv.All()
	for _, v := range all {
		ctx.Assign(v.Name, v.Value)
	}
}

func (v *ctx) All() []Variable {
	ret := []Variable{}
	inRet := func(v *Variable) bool {
		for _, lv := range ret {
			if bytes.Equal(lv.Name, v.Name) {
				return true
			}
		}
		return false
	}
	scope := v.scope
	for scope != nil {
		for _, v := range scope.vars {
			if inRet(&v) {
				continue
			}
			ret = append(ret, v)
		}
		scope = scope.p
	}
	return ret
}

func (v *ctx) Copy() Context {
	return &ctx{scope: v.scope.copy()}
}

func (v *ctx) PushScope() {
	v.scope = &scope{p: v.scope}
}

func (v *ctx) PopScope() {
	if v.scope != nil && v.scope.p != nil {
		v.scope = v.scope.p
	}
}

func (v *ctx) ValueOf(name []byte) Value {
	scope := v.scope
	for scope != nil {
		if idx := scope.indexOf(name); idx > -1 {
			return scope.vars[idx].Value
		}
		scope = scope.p
	}
	return NullValue()
}

func (v *ctx) pushMe(val Value) {
	mk := []byte{'_', 'm', 'e'}
	if idx := v.scope.indexOf(mk); idx > -1 {
		val.p = &v.scope.vars[idx].Value
	}
	v.scope.assign(mk, val)
}

func (v *ctx) popMe() {
	mk := []byte{'_', 'm', 'e'}
	if idx := v.scope.indexOf(mk); idx > -1 {
		if v.scope.vars[idx].Value.p != nil {
			v.scope.assign(mk, *v.scope.vars[idx].Value.p)
			return
		}
		v.scope.del(idx)
	}
}

func path(p string) []byte {
	return []byte(p)
}

func splitKeyAndRest(ik []byte) (k []byte, rest []byte) {
	dot := bytes.Index(ik, []byte{'.'})
	if dot < 0 {
		k = ik
		return
	}
	k = ik[0:dot]
	rest = ik[dot+1:]
	return
}

func (vs ctx) lookup(k []byte) Value {
	i, r := splitKeyAndRest(k)
	val := vs.ValueOf(i)
	if len(r) == 0 {
		return val
	}
	return val.lookup(r)
}
