# Enemy AI Behavior & Chase Logic

The enemy AI in Doomerang uses a Finite State Machine (FSM) to manage behaviors. This document outlines how enemies decide when to chase the player and, crucially, when to stop.

## State Machine Overview

Enemies cycle through three primary states defined in `systems/enemy.go`:

1.  **Patrol:** The default state. The enemy moves back and forth along a path or platform.
2.  **Chase:** The enemy has spotted the player and is actively moving towards them.
3.  **Attack:** The enemy is within range and performs a melee strike.

## The Chase Logic

### 1. Triggering the Chase
An enemy switches from **Patrol** to **Chase** when the player enters their detection radius:

```go
if distanceToPlayer <= enemy.ChaseRange {
    state.CurrentState = cfg.StateChase
}
```

### 2. The Chase Loop
While in the `StateChase`, the enemy performs the following checks every frame:

*   **Facing:** Updates direction (`1` or `-1`) to face the player.
*   **Movement:** Moves towards the player at `ChaseSpeed` if they are further away than the `StoppingDistance`.
*   **Attacking:** Transitions to `StateAttackingPunch` if the player is within `AttackRange` and the cooldown is ready.

### 3. Ending the Chase (Hysteresis)

To determine when the enemy should give up and return to **Patrol**, we use a technique called **Hysteresis**.

#### The Problem: "Flapping"
If we used the exact same distance (`ChaseRange`) to both *start* and *stop* chasing, we would encounter a glitchy behavior known as "flapping" or "flickering."

*   *Scenario:* The player stands exactly at the edge of the range (e.g., 100 pixels).
*   *Result:* The enemy would switch to Chase (player is at 99px), move slightly back, switch to Patrol (player is at 101px), and repeat rapidly. This causes the AI to twitch.

#### The Solution: Hysteresis Multiplier
We introduce a buffer zone by multiplying the chase range by a factor greater than 1 (defined in config as `HysteresisMultiplier`, typically `1.5`).

The code looks like this:

```go
if distanceToPlayer > enemy.ChaseRange * cfg.Enemy.HysteresisMultiplier {
    state.CurrentState = cfg.StatePatrol
}
```

**Example Values:**
*   **Chase Range:** 80 pixels
*   **Multiplier:** 1.5

**Resulting Behavior:**
1.  The enemy **starts** chasing when the player is **< 80 pixels** away.
2.  The enemy **continues** chasing even if the player moves to 90 or 100 pixels.
3.  The enemy **stops** chasing only when the player is **> 120 pixels** (80 * 1.5) away.

This makes the enemy feel "committed" to the chase and prevents the AI from jittering when the player is near the detection boundary.

## Configuration

These values are configured in `systems/factory/enemy.go` and `config/config.go`.
