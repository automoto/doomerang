package procgen

// ValidateGraph checks pacing rules and fixes violations in-place.
func ValidateGraph(graph *ConceptGraph) {
	if len(graph.Nodes) < 3 {
		return
	}

	validateCombatPacing(graph)
	validateVerticalPacing(graph)
}

func validateCombatPacing(graph *ConceptGraph) {
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

// isVerticalNode returns true if the node is part of a vertical section
func isVerticalNode(nt NodeType) bool {
	switch nt {
	case NodeTransitionHV, NodeTransitionVH, NodeVerticalAscent, NodeVerticalDescent, NodeVerticalCombat:
		return true
	}
	return false
}

// validateVerticalPacing ensures vertical sections are properly bookended by transitions
// and that no two vertical sections are adjacent.
func validateVerticalPacing(graph *ConceptGraph) {
	middle := graph.Nodes[1 : len(graph.Nodes)-1]

	inVertical := false
	for i := range middle {
		node := &middle[i]

		if !isVerticalNode(node.Type) {
			continue
		}

		switch node.Type {
		case NodeTransitionHV:
			if inVertical {
				node.Type = NodeTraversal
				node.Tag = TagTraversal
			} else {
				inVertical = true
			}
		case NodeTransitionVH:
			inVertical = false
		default:
			// Vertical content node without preceding transition — demote
			if !inVertical {
				node.Type = NodeTraversal
				node.Tag = TagTraversal
			}
		}
	}
}
