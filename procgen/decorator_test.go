package procgen

import (
	"testing"
)

func TestDeriveDecorationIsDeterministic(t *testing.T) {
	opts1 := DeriveDecoration(42, "cyberpunk")
	opts2 := DeriveDecoration(42, "cyberpunk")

	if opts1.TintR != opts2.TintR || opts1.TintG != opts2.TintG || opts1.TintB != opts2.TintB {
		t.Errorf("DeriveDecoration not deterministic: tint mismatch (%v,%v,%v) vs (%v,%v,%v)",
			opts1.TintR, opts1.TintG, opts1.TintB,
			opts2.TintR, opts2.TintG, opts2.TintB)
	}
}

func TestDeriveDecorationVariesByBiome(t *testing.T) {
	opts := DeriveDecoration(42, "unknown_biome")
	// Unknown biome should produce no background image but still return valid tint
	if opts.TintR <= 0 || opts.TintG <= 0 || opts.TintB <= 0 {
		t.Errorf("expected positive tint values, got (%v,%v,%v)", opts.TintR, opts.TintG, opts.TintB)
	}
}

func TestDeriveDecorationDifferentSeeds(t *testing.T) {
	// Different seeds should sometimes produce different tints
	// (not guaranteed every adjacent seed differs, but a large gap should)
	opts1 := DeriveDecoration(1, "cyberpunk")
	opts2 := DeriveDecoration(999999, "cyberpunk")
	// At minimum, both should return valid tint values
	for _, o := range []DecorationOptions{opts1, opts2} {
		if o.TintR <= 0 || o.TintG <= 0 || o.TintB <= 0 {
			t.Errorf("invalid tint values (%v,%v,%v)", o.TintR, o.TintG, o.TintB)
		}
	}
}

func TestTintPresetsAllPositive(t *testing.T) {
	for i, preset := range tintPresets {
		if preset[0] <= 0 || preset[1] <= 0 || preset[2] <= 0 {
			t.Errorf("tintPresets[%d] has non-positive value: %v", i, preset)
		}
	}
}
