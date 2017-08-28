package types

import (
	"reflect"
	"testing"
)

func TestTypeFTV(t *testing.T) {
	tests := []struct {
		t    Type
		want []string
	}{
		{
			t:    &TVar{name: "a"},
			want: []string{"a"},
		},
		{
			t:    &TInt{},
			want: nil,
		},
		{
			t:    &TFun{arg: &TInt{}, body: &TVar{name: "a"}},
			want: []string{"a"},
		},
	}
	for _, test := range tests {
		got := test.t.ftv()
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("ftv on Type: got %v, want %v", got, test.want)
		}
	}
}

func TestTypeApply(t *testing.T) {
	tests := []struct {
		t    Type
		s    Subst
		want Type
	}{
		{
			t:    &TVar{name: "a"},
			s:    Subst{"a": &TInt{}},
			want: &TInt{},
		},
		{
			t:    &TInt{},
			want: &TInt{},
		},
		{
			t:    &TFun{arg: &TVar{name: "c"}, body: &TVar{name: "b"}},
			s:    Subst{"c": &TVar{name: "a"}},
			want: &TFun{arg: &TVar{name: "a"}, body: &TVar{name: "b"}},
		},
	}
	for _, test := range tests {
		got := test.t.apply(test.s).(Type)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("apply on Type: got %v, want %v", got, test.want)
		}
	}
}

func TestComposeSubst(t *testing.T) {
	tests := []struct {
		s1   Subst
		s2   Subst
		want Subst
	}{
		{
			s1: Subst{
				"a": &TVar{"b"},
				"c": &TVar{"d"},
			},
			s2: Subst{"a": &TFun{&TVar{"a"}, &TVar{"b"}}},
			want: Subst{
				"a": &TFun{&TVar{"b"}, &TVar{"b"}},
				"c": &TVar{"d"},
			},
		},
	}
	for _, test := range tests {
		got := test.s1.compose(test.s2)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("Subst.compose: got %v, want %v", got, test.want)
		}
	}
}

func TestTypeEnvFTV(t *testing.T) {
	s := Scheme{
		vars: []string{"a"},
		t: &TFun{
			&TVar{"b"},
			&TVar{"a"},
		},
	}
	m := map[string]Scheme{"b": s}
	e := TypeEnv(m)
	got := e.ftv()
	want := []string{"b"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ftv on TypeEnv: got %v, want %v", got, want)
	}
}

func TestTypeEnvApply(t *testing.T) {
	s := Scheme{
		vars: []string{"a"},
		t: &TFun{&TVar{"a"},
			&TVar{"c"},
		},
	}
	m := Subst{
		"c": &TVar{"a"},
		"a": &TVar{"d"},
	}
	e := TypeEnv{"b": s}
	got := e.apply(m).(*TypeEnv)
	want := &TypeEnv{
		"b": Scheme{
			vars: []string{("a")},
			t: &TFun{&TVar{"a"},
				&TVar{"a"},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("apply on TypeEnv: got %v, want %v", got, want)
	}
}

func TestTINewTypeVar(t *testing.T) {
	tests := []struct {
		prefix string
		want   string
	}{
		{
			prefix: "a",
			want:   "a0",
		},
		{
			prefix: "a",
			want:   "a1",
		},
		{
			prefix: "a",
			want:   "a2",
		},
		{
			prefix: "b",
			want:   "b3",
		},
		{
			prefix: "a",
			want:   "a4",
		},
	}
	ti := TI{}
	for _, test := range tests {
		got := ti.newTypeVar(test.prefix).(*TVar).name
		if got != test.want {
			t.Errorf("TI.newTypeVar: got %v, want %v", got, test.want)
		}
	}
}

func TestTITypeInference(t *testing.T) {
	tests := []struct {
		name string
		env  TypeEnv
		expr Expr
		want Type
	}{
		{
			name: "int",
			env:  TypeEnv{},
			expr: &EInt{12},
			want: &TInt{},
		},
		{
			name: "let int",
			env:  TypeEnv{},
			expr: &ELet{"n", &EInt{12}, &EVar{"n"}},
			want: &TInt{},
		},
		{
			name: "let function",
			env:  TypeEnv{},
			expr: &ELet{"f", &EAbs{"x", &EVar{"x"}}, &EVar{"f"}},
			want: &TFun{&TVar{"a1"}, &TVar{"a1"}},
		},
		{
			name: "let function and apply it",
			env:  TypeEnv{},
			expr: &ELet{"x", &EAbs{"x", &EVar{"x"}}, &EApp{&EVar{"x"}, &EVar{"x"}}},
			want: &TFun{&TVar{"a5"}, &TVar{"a5"}},
		},
		{
			name: "let function and ignore an argument",
			env:  TypeEnv{},
			expr: &ELet{"length", &EAbs{"xs", &EInt{12}}, &EApp{&EVar{"length"}, &EVar{"length"}}},
			want: &TInt{},
		},
	}
	ti := TI{}
	for _, test := range tests {
		got := ti.typeInference(test.env, test.expr)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: typeInference = %#v; want %#v", test.name, got, test.want)
		}
	}
}

func BenchmarkTITypeInference(b *testing.B) {
	ti := TI{}
	for i := 0; i < b.N; i++ {
		ti.typeInference(TypeEnv{}, &ELet{"x", &EAbs{"x", &EVar{"x"}}, &EApp{&EVar{"x"}, &EVar{"x"}}})
	}
}
