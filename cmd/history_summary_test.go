package cmd

import "testing"

func TestCountScopeSkipped(t *testing.T) {
	cases := []struct {
		raw, kept, want int
	}{
		{10, 7, 3},
		{5, 5, 0},
		{0, 0, 0},
		{3, 5, 0}, // defensive clamp
	}
	for _, c := range cases {
		if got := countScopeSkipped(c.raw, c.kept); got != c.want {
			t.Errorf("countScopeSkipped(%d,%d) = %d, want %d", c.raw, c.kept, got, c.want)
		}
	}
}