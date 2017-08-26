package ex03

import "testing"

func TestConvertInt(t *testing.T) {
	tests := []struct {
		i int
		f float64
	}{
		{0, 0},
		{12, 12},
		{24, 24},
	}
	for n, test := range tests {
		z := ConvertInt(test.i)
		if test.f != z {
			t.Errorf("test %d expected %f, got: %f", n, test.f, z)
		}
	}
}
