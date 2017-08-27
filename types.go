package types

type Type interface {
	Type()
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
