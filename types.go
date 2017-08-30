package types

import (
	"fmt"
	"strconv"
)

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
	vars1 := f.arg.ftv()
	vars2 := f.body.ftv()
	switch {
	case len(vars1) == 0:
		return vars2
	case len(vars2) == 0:
		return vars1
	case len(vars1) < len(vars2):
		for _, v := range vars1 {
			if !contains(vars2, v) {
				vars2 = append(vars2, v)
			}
		}
		return vars2
	}
	for _, v := range vars2 {
		if !contains(vars1, v) {
			vars1 = append(vars1, v)
		}
	}
	return vars1
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
	Name string
}

type EInt struct {
	Value int
}

type EBool struct {
	Value bool
}

type EApp struct {
	Fn, Arg Expr
}

type EAbs struct {
	Param string
	Body  Expr
}

type ELet struct {
	Name string
	Bind Expr
	Body Expr
}

type EIf struct {
	cond Expr
	e1   Expr
	e2   Expr
}

func (e *EVar) Expr()  {}
func (e *EInt) Expr()  {}
func (e *EBool) Expr() {}
func (e *EApp) Expr()  {}
func (e *EAbs) Expr()  {}
func (e *ELet) Expr()  {}
func (e *EIf) Expr()   {}

type Subst map[string]Type

// compose composes two Subst.
// Note that this method can update a receiver's value.
func (s Subst) compose(s0 Subst) Subst {
	if len(s) == 0 {
		return s0
	}
	m := make(Subst, len(s0))
	for k, v := range s {
		m[k] = v
	}
	for k, v := range s0 {
		m[k] = v.apply(s).(Type)
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
	m := make(Subst, len(sub))
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

func (env TypeEnv) generalize(t Type) Scheme {
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

func (env TypeEnv) ftv() []string {
	ret := make([]string, 0, len(env))
	for _, s := range env {
		for _, v := range s.ftv() {
			if !contains(ret, v) {
				ret = append(ret, v)
			}
		}
	}
	return ret
}

func (env TypeEnv) apply(s Subst) Types {
	for k, v := range env {
		env[k] = *v.apply(s).(*Scheme)
	}
	return env
}

type TI struct {
	supply uint
}

func (ti *TI) newTypeVar(s string) Type {
	n := ti.supply
	ti.supply++
	return &TVar{name: s + strconv.Itoa(int(n))}
}

func (ti *TI) instantiate(s Scheme) Type {
	m := make(Subst, len(s.vars))
	for _, v := range s.vars {
		m[v] = ti.newTypeVar("a")
	}
	return s.t.apply(m).(Type)
}

func (ti *TI) varBind(u string, t Type) (Subst, error) {
	if x, ok := t.(*TVar); ok && x.name == u {
		return nil, nil
	}
	if !contains(t.ftv(), u) {
		return Subst{u: t}, nil
	}
	return nil, fmt.Errorf("occur check fails: %s vs. %v", u, t)
}

func (ti *TI) mgu(t1, t2 Type) (Subst, error) {
	switch x := t1.(type) {
	case *TFun:
		if y, ok := t2.(*TFun); ok {
			s1, err := ti.mgu(x.arg, y.arg)
			if err != nil {
				return nil, err
			}
			s2, err := ti.mgu(x.body.apply(s1).(Type), y.body.apply(s1).(Type))
			if err != nil {
				return nil, err
			}
			return s1.compose(s2), nil
		}
		if v, ok := t2.(*TVar); ok {
			return ti.varBind(v.name, t1)
		}
	case *TVar:
		return ti.varBind(x.name, t2)
	case *TInt:
		switch y := t2.(type) {
		case *TVar:
			return ti.varBind(y.name, t1)
		case *TInt:
			return nil, nil
		}
	case *TBool:
		switch y := t2.(type) {
		case *TVar:
			return ti.varBind(y.name, t1)
		case *TBool:
			return nil, nil
		}
	}
	return nil, fmt.Errorf("types do not unify: %#v vs. %#v", t1, t2)
}

func (ti *TI) ti(env TypeEnv, expr Expr) (Subst, Type, error) {
	switch e := expr.(type) {
	case *EVar:
		sigma, ok := env[e.Name]
		if !ok {
			return nil, nil, fmt.Errorf("unbound variable: %s", e.Name)
		}
		return nil, ti.instantiate(sigma), nil
	case *EInt:
		return nil, &TInt{}, nil
	case *EBool:
		return nil, &TBool{}, nil
	case *EApp:
		tv := ti.newTypeVar("a")
		s1, t1, err := ti.ti(env, e.Fn)
		if err != nil {
			return nil, nil, err
		}
		s2, t2, err := ti.ti(env.apply(s1).(TypeEnv), e.Arg)
		if err != nil {
			return nil, nil, err
		}
		s3, err := ti.mgu(t1.apply(s2).(Type), &TFun{arg: t2, body: tv})
		if err != nil {
			return nil, nil, err
		}
		s := s3.compose(s2)
		return s.compose(s1), tv.apply(s3).(Type), nil
	case *EAbs:
		tv := ti.newTypeVar("a")
		x, ok := env[e.Param]
		env[e.Param] = Scheme{t: tv}
		s1, t1, err := ti.ti(env, e.Body)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			env[e.Param] = x
		} else {
			delete(env, e.Param)
		}
		return s1, &TFun{arg: tv.apply(s1).(Type), body: t1}, nil
	case *ELet:
		s1, t1, err := ti.ti(env, e.Bind)
		if err != nil {
			return nil, nil, err
		}
		t := env.apply(s1).(TypeEnv).generalize(t1)
		x, ok := env[e.Name]
		env[e.Name] = t
		s2, t2, err := ti.ti(env.apply(s1).(TypeEnv), e.Body)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			env[e.Name] = x
		} else {
			delete(env, e.Name)
		}
		return s1.compose(s2), t2, nil
	case *EIf:
		s1, t1, err := ti.ti(env, e.cond)
		if err != nil {
			return nil, nil, err
		}
		s2, err := ti.mgu(t1.apply(s1).(Type), &TBool{})
		if err != nil {
			return nil, nil, err
		}
		s3, t2, err := ti.ti(env.apply(s2).(TypeEnv), e.e1)
		if err != nil {
			return nil, nil, err
		}
		s4, t3, err := ti.ti(env.apply(s3).(TypeEnv), e.e2)
		if err != nil {
			return nil, nil, err
		}
		s5, err := ti.mgu(t2.apply(s4).(Type), t3.apply(s4).(Type))
		if err != nil {
			return nil, nil, err
		}
		s := s5.compose(s4).compose(s3).compose(s2).compose(s1)
		return s, t3.apply(s5).(Type), nil
	}
	panic("unreachable")
}

func (ti *TI) TypeInference(env TypeEnv, expr Expr) (Type, error) {
	s, t, err := ti.ti(env, expr)
	if err != nil {
		return nil, err
	}
	return t.apply(s).(Type), nil
}
