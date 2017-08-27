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
