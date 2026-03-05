package procgen_test

import (
	"testing"

	"github.com/automoto/doomerang/procgen"
)

func TestDifficultyRange(t *testing.T) {
	for total := 2; total <= 15; total++ {
		for pos := 0; pos < total; pos++ {
			d := procgen.DifficultyAtNode(pos, total)
			if d < 1 || d > 5 {
				t.Errorf("DifficultyAtNode(%d, %d) = %d, want [1,5]", pos, total, d)
			}
		}
	}
}

func TestDifficultySCurve(t *testing.T) {
	total := 12
	first := procgen.DifficultyAtNode(0, total)
	mid := procgen.DifficultyAtNode(total/2, total)
	last := procgen.DifficultyAtNode(total-1, total)

	if first > 2 {
		t.Errorf("start difficulty too high: %d", first)
	}
	if mid < 2 || mid > 4 {
		t.Errorf("mid difficulty unexpected: %d", mid)
	}
	if last < 4 {
		t.Errorf("end difficulty too low: %d", last)
	}
}

func TestDifficultyMonotonic(t *testing.T) {
	total := 12
	prev := procgen.DifficultyAtNode(0, total)
	for i := 1; i < total; i++ {
		curr := procgen.DifficultyAtNode(i, total)
		if curr < prev {
			t.Errorf("difficulty decreased at position %d: %d -> %d", i, prev, curr)
		}
		prev = curr
	}
}
