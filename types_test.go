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
