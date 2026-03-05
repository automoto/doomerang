package procgen

// ValidateGraph checks pacing rules and fixes violations in-place.
func ValidateGraph(graph *ConceptGraph) {
	if len(graph.Nodes) < 3 {
		return
	}

	// Only validate middle nodes (skip start at 0 and exit at end)
	middle := graph.Nodes[1 : len(graph.Nodes)-1]

	combatStreak := 0
	combatsSinceBreak := 0

	for i := range middle {
		node := &middle[i]

		switch node.Type {
		case NodeCombat, NodeArena:
			combatStreak++
			combatsSinceBreak++

			// Enforce: max 2 consecutive combat
			if combatStreak > 2 {
				node.Type = NodeTraversal
				node.Tag = TagTraversal
				combatStreak = 0
			}

			// Enforce: break room after 3 combats
			if combatsSinceBreak > 3 {
				node.Type = NodeBreakRoom
				node.Tag = TagBreak
				combatStreak = 0
				combatsSinceBreak = 0
			}

		case NodeBreakRoom:
			combatStreak = 0
			combatsSinceBreak = 0

		default:
			combatStreak = 0
		}
	}
}
