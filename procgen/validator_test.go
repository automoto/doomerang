package procgen_test

import (
	"math/rand"
	"testing"

	"github.com/automoto/doomerang/procgen"
)

func TestValidatorBasicLevel(t *testing.T) {
	chunks := loadTestChunks(t)

	gen := procgen.NewChunkGenerator(42)
	result, err := gen.Generate(chunks, 3)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	validator := procgen.NewValidator()
	vr := validator.Validate(result)

	if vr.PlatformCount == 0 {
		t.Error("expected platforms to be discovered")
	}
	if !vr.Solvable {
		t.Errorf("basic level should be solvable, %d unreachable platforms", len(vr.Unreachable))
	}
}

func TestValidatorMultipleSeeds(t *testing.T) {
	chunks := loadTestChunks(t)
	validator := procgen.NewValidator()

	for seed := int64(0); seed < 20; seed++ {
		gen := procgen.NewChunkGenerator(seed)
		result, err := gen.Generate(chunks, 5)
		if err != nil {
			continue
		}

		vr := validator.Validate(result)
		if !vr.Solvable {
			t.Logf("seed %d: level not solvable (%d unreachable of %d platforms)",
				seed, len(vr.Unreachable), vr.PlatformCount)
		}
	}
}

func TestValidatorGraphLevel(t *testing.T) {
	chunks := loadTestChunks(t)
	validator := procgen.NewValidator()

	for seed := int64(0); seed < 10; seed++ {
		rng := rand.New(rand.NewSource(seed))
		graph := procgen.GenerateGraph(rng, 5, []string{"cyberpunk"})
		procgen.ValidateGraph(graph)

		gen := procgen.NewChunkGenerator(seed)
		result, err := gen.GenerateFromGraph(chunks, graph)
		if err != nil {
			continue
		}

		vr := validator.Validate(result)
		if vr.PlatformCount == 0 {
			t.Errorf("seed %d: no platforms found", seed)
		}
	}
}

func TestValidateAndRemediate(t *testing.T) {
	chunks := loadTestChunks(t)
	rng := rand.New(rand.NewSource(42))
	graph := procgen.GenerateGraph(rng, 5, []string{"cyberpunk"})
	procgen.ValidateGraph(graph)

	gen := procgen.NewChunkGenerator(42)
	result, err := procgen.ValidateAndRemediate(gen, chunks, graph, 5)
	if err != nil {
		t.Fatalf("ValidateAndRemediate failed: %v", err)
	}

	if len(result.PlacedChunks) == 0 {
		t.Error("expected placed chunks")
	}
	if result.TotalWidth <= 0 {
		t.Error("expected positive total width")
	}
}

func TestCanReachAdjacentFloors(t *testing.T) {
	// Two chunks side by side with same floor height should be walkable
	chunks := loadTestChunks(t)

	gen := procgen.NewChunkGenerator(42)
	// Use just start + exit (minimal level)
	result, err := gen.Generate(chunks, 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	validator := procgen.NewValidator()
	vr := validator.Validate(result)

	// With standard connection heights, adjacent floors should be reachable
	if !vr.Solvable {
		t.Errorf("simple level should be solvable, unreachable: %v", vr.Unreachable)
	}
}
