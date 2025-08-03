# Boomerang Weapon: Game Mechanic & Technical Specification

This document outlines the core mechanics, behavior, and feel for the high-powered boomerang weapon in the game. It is based on the detailed technical design document and specifies the desired implementation choices.

## 1. Overview & Core Feel

The boomerang is a versatile weapon with a dual-mode feel based on player input. Its behavior is designed to be intuitive for action-platforming while offering tactical depth.

-   **Quick Throw (Tap):** A fast, light, and agile attack. It deals standard damage and allows the player to react quickly.
-   **Charged Throw (Hold):** A "bone-crushing" power attack. The player holds the throw button to charge it. This version is slower, heavier, deals significantly more damage, and should be accompanied by more impactful audio-visual feedback.

## 2. Throw Mechanic (Outbound Path)

The boomerang's outward path should be predictable and feel connected to the player's actions.

-   **Path Type:** **Parametric Arc (Simulated Gravity)**. The boomerang will follow a standard projectile motion curve influenced by an adjustable gravity constant.
-   **Initial Velocity:** The throw's initial velocity vector will be calculated from a base speed and angle, but will also incorporate a fraction of the player's current momentum to feel responsive.
-   **Charge Influence:**
    -   *Quick Throw:* Lower initial velocity.
    -   *Charged Throw:* Higher initial velocity.

## 3. Player Control In-Flight

The player retains a key piece of strategic control over the boomerang after it has been thrown.

-   **Early Recall:** The player can press the throw button a second time while the boomerang is in its `OUTBOUND` state to immediately interrupt its path and trigger the `INBOUND` (return) state.

## 4. Return Mechanic (Inbound Path)

The boomerang's return is aggressive and direct.

-   **Homing Type:** **Simple Vector-Based Homing**. The boomerang will calculate a straight-line vector to the player's current position each frame and move directly towards it. It will not use smoothing or interpolation, resulting in sharp, immediate turns to track the player.

## 5. Collision Logic

The boomerang's interaction with the game world is state-dependent and designed for offensive pressure.

-   **Enemy Collision:**
    -   **Behavior:** The boomerang is **Piercing**. It will not stop or return after hitting an enemy.
    -   **Multi-hit Logic:** To prevent a single throw from hitting one enemy multiple times, the boomerang must maintain a list of enemies it has already damaged. It can damage an enemy once on the `OUTBOUND` path and once again on the `INBOUND` path.
    -   **Damage:** Damage dealt will depend on whether it was a *Quick Throw* or a *Charged Throw*.

-   **Environment Collision (Walls, Floors):**
    -   **Behavior:** **Stop and Return**. Upon colliding with any static world geometry, the boomerang's path is interrupted, and it will immediately transition to the `INBOUND` state. It does not ricochet.

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
-   **`ProjectileComponent`**: Holds data specific to projectile motion, including initial velocity, time in flight, and a gravity scale.
-   **`BoomerangComponent`**: Contains boomerang-specific attributes:
    -   `maxRange`: Maximum travel distance.
    -   `returnSpeed`: Speed during the `INBOUND` state.
    -   `chargeLevel`: A value indicating if it was a quick or charged throw.
    -   `hitEnemies`: A list to track entities already damaged to manage piercing logic.
-   **`CollisionComponent`**: Defines the hitbox shape, size, and collision mask.
-   **`RenderComponent`**: Contains data for rendering, like sprite and trail effect IDs.

### 6.3 Core Systems (Logic)

-   **`PlayerInputSystem`**: Reads player input. If the throw button is pressed, it can add a `WantsToThrowComponent` to the player entity.
-   **`BoomerangThrowSystem`**: Looks for players with `WantsToThrowComponent`. Manages the logic for charging and throwing. When a throw occurs, it creates the boomerang entity, attaches the necessary components (including `OutboundState`), and sets its initial velocity based on charge level and player momentum.
-   **`ProjectileSystem`**: Queries for entities with `ProjectileComponent`, `VelocityComponent`, `TransformComponent`, and `OutboundState`. It updates the entity's position based on the parametric arc calculation each frame.
-   **`HomingSystem`**: Queries for entities with `BoomerangComponent`, `VelocityComponent`, `TransformComponent`, and `InboundState`. It updates the entity's velocity to move it directly towards the player.
-   **`CollisionSystem`**: Manages all collision detection. When a boomerang collides with something:
    -   *vs. Enemy*: Checks the `hitEnemies` list, applies damage, and updates the list.
    -   *vs. Wall*: Removes the `OutboundState` tag and adds the `InboundState` tag.
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
