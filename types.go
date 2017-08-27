package types

type Types interface {
	ftv() []string
	apply(Subst) Types
}

type Type interface {
	Type()
	Types
}

type Var struct {
	name string
}

type Int struct{}

type Bool struct{}

type Fun struct {
	arg, body Type
}

func (v *Var) Type()  {}
func (i *Int) Type()  {}
func (b *Bool) Type() {}
func (f *Fun) Type()  {}

func (v *Var) ftv() []string {
	return []string{v.name}
}

func (i *Int) ftv() []string {
	return nil
}

func (b *Bool) ftv() []string {
	return nil
}

func (f *Fun) ftv() []string {
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

func (v *Var) apply(s Subst) Types {
	if t, ok := s[v.name]; ok {
		return t
	}
	return v
}

func (i *Int) apply(s Subst) Types {
	return i
}

func (b *Bool) apply(s Subst) Types {
	return b
}

func (f *Fun) apply(s Subst) Types {
	return &Fun{
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
