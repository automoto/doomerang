package procgen_test

import (
	"math/rand"
	"testing"

	"github.com/automoto/doomerang/procgen"
)

func TestValidatorBasicLevel(t *testing.T) {
	chunks := loadTestChunks(t)

	assembler := procgen.NewAssembler(42)
	result, err := assembler.Assemble(chunks, 3)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
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
		assembler := procgen.NewAssembler(seed)
		result, err := assembler.Assemble(chunks, 5)
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

		assembler := procgen.NewAssembler(seed)
		result, err := assembler.AssembleFromGraph(chunks, graph)
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

	assembler := procgen.NewAssembler(42)
	result, err := procgen.ValidateAndRemediate(assembler, chunks, graph, 5)
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

	assembler := procgen.NewAssembler(42)
	// Use just start + exit (minimal level)
	result, err := assembler.Assemble(chunks, 1)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	validator := procgen.NewValidator()
	vr := validator.Validate(result)

	// With standard connection heights, adjacent floors should be reachable
	if !vr.Solvable {
		t.Errorf("simple level should be solvable, unreachable: %v", vr.Unreachable)
	}
}
