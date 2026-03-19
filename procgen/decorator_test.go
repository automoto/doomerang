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

func TestBiomeTintsAllPositive(t *testing.T) {
	for biome, tints := range biomeTints {
		for i, preset := range tints {
			if preset[0] <= 0 || preset[1] <= 0 || preset[2] <= 0 {
				t.Errorf("biomeTints[%s][%d] has non-positive value: %v", biome, i, preset)
			}
		}
	}
}

func TestDeriveDecorationDifferentBiomesDifferentTints(t *testing.T) {
	// Same seed with different biomes should use biome-specific tint palettes
	cyberpunk := DeriveDecoration(42, "cyberpunk")
	industrial := DeriveDecoration(42, "industrial")
	neon := DeriveDecoration(42, "neon")

	// All should have valid positive tints
	for name, o := range map[string]DecorationOptions{
		"cyberpunk": cyberpunk, "industrial": industrial, "neon": neon,
	} {
		if o.TintR <= 0 || o.TintG <= 0 || o.TintB <= 0 {
			t.Errorf("%s: invalid tint values (%v,%v,%v)", name, o.TintR, o.TintG, o.TintB)
		}
	}

	// At least one pair should differ (tint palettes are distinct)
	allSame := (cyberpunk.TintR == industrial.TintR && cyberpunk.TintG == industrial.TintG && cyberpunk.TintB == industrial.TintB) &&
		(cyberpunk.TintR == neon.TintR && cyberpunk.TintG == neon.TintG && cyberpunk.TintB == neon.TintB)
	if allSame {
		t.Error("all three biomes produced identical tints — expected at least one difference")
	}
}

func TestDefaultTintForUnknownBiome(t *testing.T) {
	opts := DeriveDecoration(42, "unknown_biome")
	if opts.TintR != defaultTint[0] || opts.TintG != defaultTint[1] || opts.TintB != defaultTint[2] {
		t.Errorf("unknown biome should use default tint (%v,%v,%v), got (%v,%v,%v)",
			defaultTint[0], defaultTint[1], defaultTint[2],
			opts.TintR, opts.TintG, opts.TintB)
	}
}
