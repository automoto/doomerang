# Boomerang Weapon: Game Mechanic & Technical Specification

This document outlines the core mechanics, behavior, and feel for the high-powered boomerang weapon in the game. It is based on the detailed technical design document and specifies the desired implementation choices.

## 1. Overview & Core Feel

The boomerang is a versatile weapon controlled by the **Spacebar**. Its behavior is designed to be intuitive for action-platforming while offering tactical depth.

-   **Input:** Press `Spacebar` to charge, release to throw.
-   **Player Movement:** The player will decelerate and stop moving while charging and performing the throw animation, similar to other melee attacks.
-   **Quick Throw (Tap):** A fast, light attack with a shorter maximum range.
-   **Charged Throw (Hold):** A "bone-crushing" power attack. The player holds the throw button to charge it. This version is slower, heavier, deals significantly more damage, and has a longer maximum range.

## 2. Throw Mechanic (Outbound Path)

The boomerang's outward path should be predictable and feel connected to the player's actions.

-   **Path Type:** **Parametric Arc (Simulated Gravity)**. The boomerang will follow a standard projectile motion curve influenced by an adjustable gravity constant.
-   **Range & Velocity:**
    -   *Quick Throw:* Lower initial velocity, shorter max distance.
    -   *Charged Throw:* Higher initial velocity, longer max distance.
-   **Visuals:** Uses `throw.png` for the player animation and `boom_green.png` for the boomerang sprite.



The player **does not** have manual control over the return. The boomerang returns automatically based on specific conditions.

-   **Max Range Return:** The boomerang transitions to the `INBOUND` state automatically when it reaches its maximum distance (determined by the charge level).

## 4. Return Mechanic (Inbound Path)

The boomerang's return is aggressive and direct.

-   **Homing Type:** **Simple Vector-Based Homing**. The boomerang will calculate a straight-line vector to the player's current position each frame and move directly towards it. It will not use smoothing or interpolation, resulting in sharp, immediate turns to track the player.

## 5. Collision Logic

The boomerang's interaction with the game world is state-dependent and designed for offensive pressure.

-   **Enemy Collision:**
    -   **Behavior:** The boomerang is **Piercing**. It will not stop or return immediately after hitting an enemy.
    -   **Short Return Rule:** Upon hitting an enemy, the boomerang's remaining range is **truncated**. It will travel a short "pierce distance" beyond the enemy and then automatically transition to the `INBOUND` state. This prevents the weapon from flying too far off-screen after a successful hit.
    -   **Multi-hit Logic:** To prevent a single throw from hitting one enemy multiple times, the boomerang must maintain a list of enemies it has already damaged. It can damage an enemy once on the `OUTBOUND` path and once again on the `INBOUND` path.
    -   **Damage:** Damage dealt will depend on whether it was a *Quick Throw* or a *Charged Throw*.

-   **Environment Collision (Walls, Floors):**
    -   **Behavior:** **Immediate Stop and Return**. Upon colliding with any static world geometry, the boomerang's path is interrupted immediately, and it will immediately transition to the `INBOUND` state. It does not ricochet.

-   **Player Collision:**
    -   **Behavior:** This collision is only active during the `INBOUND` state. When the boomerang touches the player, it is "caught." This transitions the boomerang back to the `HELD` state and resets the player's ability to throw again.

## 6. Architecture: Entity Component System (ECS)

The boomerang will be implemented using an ECS architecture. It will be an **entity** defined by the following components and managed by corresponding systems.

### 6.1 State Management

A Finite State Machine (FSM) will be implemented using **State Tag Components**. The boomerang entity will have one of the following empty "tag" components attached at any time to represent its current state.

-   `HeldState`
-   `OutboundState`
-   `InboundState`

### 6.2 Core Components (Data)

-   **`TransformComponent`**: Stores position and rotation.
-   **`VelocityComponent`**: Stores current linear and angular velocity.
-   **`PhysicsComponent`**: Standard physics integration (gravity, velocity).
-   **`BoomerangComponent`**: Contains boomerang-specific attributes:
    -   `State`: Current state (Outbound/Inbound).
    -   `Owner`: Reference to the player entity.
    -   `MaxRange`: Maximum travel distance (determined by charge).
    -   `DistanceTraveled`: Tracked frame-by-frame.
    -   `PierceDistance`: Fixed distance to travel after hitting an enemy.
    -   `ReturnSpeed`: Speed during the `INBOUND` state.
    -   `HitEnemies`: A list to track entities already damaged to manage piercing logic.
-   **`CollisionComponent`**: Defines the hitbox shape, size, and collision mask.
-   **`RenderComponent`**: Contains data for rendering, like sprite and trail effect IDs.

### 6.3 Core Systems (Logic)

-   **`PlayerInputSystem`**: Reads player input (`Spacebar`). Manages logic for charging (with player deceleration) and throwing.
-   **`Factory`**: Handles the creation of the boomerang entity with the correct sprite (`boom_green.png`) and components.
-   **`BoomerangSystem`**:
    -   **Outbound:** Updates position based on parametric arc. Checks if `DistanceTraveled >= MaxRange` to trigger `Inbound`.
    -   **Inbound:** Updates velocity to home in on the player.
    -   **Sprite:** Handles rotation logic.
-   **`CollisionSystem`**: Manages all collision detection. When a boomerang collides with something:
    -   *vs. Enemy*: Checks the `hitEnemies` list, applies damage, and triggers the **Short Return Rule**.
    -   *vs. Wall*: Removes the `OutboundState` tag and adds the `InboundState` tag (Immediate Return).
    -   *vs. Player (Inbound only)*: Destroys the boomerang entity.
-   **`RenderSystem`**: Draws all entities with `RenderComponent` and `TransformComponent`.

## 7. "Juice" - Audio-Visual Polish

Feedback is critical to making the two throw types feel distinct and satisfying.

-   **Charged Throw:**
    -   **Visuals:** A significant particle trail, a charge-up effect on the player, and a impactful visual effect on enemy hit. A brief **Hit Stop** (1-3 frames) and subtle **Screen Shake** should occur on enemy impact.
    -   **Audio:** A powerful "thwump" throw sound, a heavy impact sound on hit.

-   **Quick Throw:**
    -   **Visuals:** A lighter trail and less dramatic impact effects.
    -   **Audio:** A sharper, quicker "shing" or "whoosh" on throw, and a lighter impact sound.

-   **Universal Sounds:** A flight loop sound (pitch-modulated by velocity) and a clean "catch" sound are required to complete the gameplay loop.