package types

type Types interface {
	ftv() []string
	apply(Subst) Types
}

type Type interface {
	Type()
	Types
}

type TVar struct {
	name string
}

type TInt struct{}

type TBool struct{}

type TFun struct {
	arg, body Type
}

func (v *TVar) Type()  {}
func (i *TInt) Type()  {}
func (b *TBool) Type() {}
func (f *TFun) Type()  {}

func (v *TVar) ftv() []string {
	return []string{v.name}
}

func (i *TInt) ftv() []string {
	return nil
}

func (b *TBool) ftv() []string {
	return nil
}

func (f *TFun) ftv() []string {
	vars := f.arg.ftv()
	for _, v := range f.body.ftv() {
		if !contains(vars, v) {
			vars = append(vars, v)
		}
	}
	return vars
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if x == y {
			return true
		}
	}
	return false
}

func (v *TVar) apply(s Subst) Types {
	if t, ok := s[v.name]; ok {
		return t
	}
	return v
}

func (i *TInt) apply(s Subst) Types {
	return i
}

func (b *TBool) apply(s Subst) Types {
	return b
}

func (f *TFun) apply(s Subst) Types {
	return &TFun{
		arg:  f.arg.apply(s).(Type),
		body: f.body.apply(s).(Type),
	}
}

type Expr interface {
	Expr()
}

type EVar struct {
	name string
}

type EInt struct {
	value int
}

type EBool struct {
	value bool
}

type EApp struct {
	fn, arg Expr
}

type EAbs struct {
	param string
	expr  Expr
}

type ELet struct {
	name string
	bind Expr
	expr Expr
}

func (e *EVar) Expr()  {}
func (e *EInt) Expr()  {}
func (e *EBool) Expr() {}
func (e *EApp) Expr()  {}
func (e *EAbs) Expr()  {}
func (e *ELet) Expr()  {}

type Subst map[string]Type

func (s *Subst) compose(s0 Subst) Subst {
	m := make(map[string]Type, len(s0))
	for k, v := range s0 {
		m[k] = v.apply(*s).(Type)
	}
	for k, v := range *s {
		if _, found := m[k]; !found {
			m[k] = v
		}
	}
	return m
}

type Scheme struct {
	vars []string
	t    Type
}

func (s *Scheme) ftv() []string {
	list := s.t.ftv()
	ret := make([]string, 0, len(list))
	for _, x := range list {
		if !contains(s.vars, x) {
			ret = append(ret, x)
		}
	}
	return ret
}

func (s *Scheme) apply(sub Subst) Types {
	m := make(map[string]Type, len(sub))
	for k, v := range sub {
		if !contains(s.vars, k) {
			m[k] = v
		}
	}
	return &Scheme{
		vars: s.vars,
		t:    s.t.apply(m).(Type),
	}
}

type TypeEnv map[string]Scheme

func (env *TypeEnv) generalize(t Type) Scheme {
	tftv := t.ftv()
	eftv := env.ftv()
	vars := make([]string, 0, len(tftv))
	for _, x := range tftv {
		if !contains(eftv, x) {
			vars = append(vars, x)
		}
	}
	return Scheme{
		vars: vars,
		t:    t,
	}
}

func (env *TypeEnv) ftv() []string {
	ret := make([]string, len(*env))
	for _, s := range *env {
		for _, v := range s.ftv() {
			if !contains(ret, v) {
				ret = append(ret, v)
			}
		}
	}
	return ret
}

func (env *TypeEnv) apply(s Subst) Types {
	for k, v := range *env {
		(*env)[k] = *v.apply(s).(*Scheme)
	}
	return env
}
