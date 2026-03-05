package procgen

import "math"

// DifficultyAtNode returns a difficulty value (1-5) for a given node position
// using an S-curve (logistic function) across the run.
// position is 0-based index, total is the number of nodes.
func DifficultyAtNode(position, total int) int {
	if total <= 1 {
		return 1
	}

	// Normalize position to 0..1
	t := float64(position) / float64(total-1)

	// Logistic S-curve: steep ramp in the middle, plateaus at ends
	// k controls steepness, x0 is the midpoint
	k := 8.0
	x0 := 0.5
	s := 1.0 / (1.0 + math.Exp(-k*(t-x0)))

	// Map from [0,1] to [1,5]
	diff := 1.0 + s*4.0
	return int(math.Round(diff))
}
