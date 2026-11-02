package transport

import "testing"

func TestNormalizeFragSize(t *testing.T) {
	for _, tc := range []struct{ size, def, want uint16 }{
		{0, 4280, 4280},
		{1, 4280, 1432},
		{1431, 4280, 1432},
		{1432, 4280, 1432},
		{1433, 4280, 1433},
		{0xFFFF, 4280, 0xFFFF},
	} {
		if got := NormalizeFragSize(tc.size, tc.def); got != tc.want {
			t.Errorf("NormalizeFragSize(%d, %d) = %d, want %d", tc.size, tc.def, got, tc.want)
		}
	}
}
