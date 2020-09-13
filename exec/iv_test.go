package exec

import (
	"reflect"
	"testing"
)

func TestReduce_Array_RHS_OK(t *testing.T) {
	cases := []struct {
		name   string
		array  []ZnValue
		index  string
		expect ZnValue
	}{
		{
			"normal RHS value",
			[]ZnValue{
				NewZnString("64"),
				NewZnString("128"),
			},
			"0",
			NewZnString("64"),
		},
		{
			"normal RHS value #2",
			[]ZnValue{
				NewZnBool(true),
				NewZnString("Hello World"),
				NewZnArray([]ZnValue{
					NewZnString("A"),
					NewZnString("B"),
				}),
			},
			"2",
			NewZnArray([]ZnValue{
				NewZnString("A"),
				NewZnString("B"),
			}),
		},
	}

	for _, tt := range cases {
		fakeInput := NewZnString("fake_input")
		t.Run(tt.name, func(t *testing.T) {
			decimal, _ := NewZnDecimal(tt.index)
			ctx := NewContext()
			scope := NewRootScope()
			iv := ZnArrayIV{NewZnArray(tt.array), decimal}

			v, err := iv.Reduce(ctx, scope, fakeInput, false)
			if err != nil {
				t.Errorf("reduce() should have no error - but error: %s occured", err.Error())
				return
			}

			if !reflect.DeepEqual(v, tt.expect) {
				t.Errorf("not same: expect=%v, reduced=%v", v.String(), tt.expect.String())
			}
		})
	}
}

// TODO: add more testcases
