# 2D Platformer Math and Physics Concepts

This document provides a summary of the fundamental math and physics concepts required to build a 2D platformer game, using this project as a reference.

## 1. Core Game Loop and Physics

The game runs on a fixed-step game loop provided by the Ebiten game engine. The main update logic for the player is managed by an Entity Component System (ECS). In `scenes/platformer.go`, we can see the `UpdatePlayer` system is registered, which gets called every frame.

The `UpdatePlayer` function in `systems/player.go` orchestrates player behavior by calling a sequence of functions:
1.  `handlePlayerInput`: Reads keyboard/gamepad input to change the player's intent.
2.  `applyPlayerPhysics`: Updates the player's velocity based on acceleration, gravity, and friction.
3.  `resolvePlayerCollisions`: Moves the player and resolves any collisions with the environment.

This approach separates input, physics simulation, and collision response, which is a common and effective pattern.

## 2. Vectors and Player State

While the term "vector" is useful, in this codebase, physics-related properties are often stored as individual fields.

-   **Position**: The `resolv.Object` in `components/object.go` stores the player's `X` and `Y` coordinates in the game world. This is the authoritative source for the player's location.
-   **Velocity**: The `PlayerData` component in `components/player.go` stores `SpeedX` and `SpeedY`. These represent the velocity vector's components.
-   **State**: `PlayerData` also holds crucial state information like `OnGround`, `WallSliding`, and `FacingRight`, which dictate how the player interacts with the world.
-   **Camera Position**: The camera uses a true vector, `math.Vec2`, from the `donburi` library to store its position (`components/camera.go`).

## 3. Movement and Controls

All movement logic is defined within `systems/player.go`.

### Horizontal Movement

Horizontal movement is handled in `handlePlayerInput` and `applyPlayerPhysics`.
-   **Acceleration**: Pressing left or right adds the `playerAccel` constant to `player.SpeedX`.
-   **Friction**: When there is no input, `playerFriction` is subtracted from `player.SpeedX` until it reaches zero.
-   **Max Speed**: `player.SpeedX` is clamped to `playerMaxSpeed` to prevent unlimited acceleration.

### Gravity

A constant downward acceleration, `playerGravity`, is applied to `player.SpeedY` every frame in `applyPlayerPhysics`. When the player is wall-sliding, gravity's effect is reduced (`player.SpeedY` is capped at `1`) to simulate sliding down the wall instead of falling at full speed.

### Jumping

Jumping is initiated in `handlePlayerInput` by setting `player.SpeedY` to a negative value (`-playerJumpSpd`). The logic distinguishes between several jump types:
-   **Ground Jump**: Can only be performed if `player.OnGround` is not `nil`.
-   **Wall Jump**: Can be performed if `player.WallSliding` is not `nil`. This also imparts a horizontal boost away from the wall.
-   **Dropping Through Platforms**: If the player is on an object with the "platform" tag, pressing down and jump allows them to fall through it by temporarily storing the platform in `player.IgnorePlatform`.

## 4. Collision System

This project uses the `resolv` library for collision detection and resolution.

-   **Space**: In `scenes/platformer.go`, a `resolv.Space` is created. This is a spatial hash that organizes all physical objects to make collision checks efficient. All solid objects, including the player, are added to this space.
-   **Collision Check**: In `resolvePlayerCollisions` (`systems/player.go`), `playerObject.Check(dx, dy, ...tags)` is used. This function checks if moving the player by `(dx, dy)` would result in a collision with any objects that have the specified tags (e.g., "solid", "platform").
-   **Resolution**: The resolution logic is separated into a horizontal pass and a vertical pass.
    1.  The horizontal movement (`dx`) is checked and applied. If a collision occurs with a "solid" wall, movement is stopped, `SpeedX` is zeroed, and `WallSliding` state may be triggered.
    2.  The vertical movement (`dy`) is checked and applied. This pass is more complex, handling interactions with ramps, platforms, and solid ground. On a downward collision, `player.OnGround` is set, and `SpeedY` is zeroed.

## 5. Camera Control

The camera smoothly follows the player. This is implemented in `systems/camera.go`. The logic uses linear interpolation (lerp) to create a soft, elastic-like camera movement.

`camera.Position.X += (playerObject.X - camera.Position.X) * 0.1`
`camera.Position.Y += (playerObject.Y - camera.Position.Y) * 0.1`

This code moves the camera 10% of the distance towards the player each frame, which results in a smooth follow effect instead of a rigid lock.

## 6. Animation

Player animation is tied to player state.

-   **Animation Component**: `components/animation.go` defines `AnimationData`, which holds a map of all possible animations (e.g., "run", "jump", "stand") and a reference to the `CurrentAnimation`.
-   **State-Driven Animation**: The `updatePlayerAnimation` function in `systems/player.go` is a state machine that sets the current animation based on the player's state (`OnGround`, `SpeedX`). For example, if the player is not on the ground, the "jump" animation is played.
-   **Rendering**: `DrawPlayer` uses the `CurrentAnimation` to determine which frame of which spritesheet to draw. It also handles flipping the sprite based on `player.FacingRight`.

## Further Reading and Citations

-   **Basic Physics and Platformer Movement:**
    -   Gaffer on Games - "Integration Basics": A fantastic series on game physics.
        -   *Link*: https://gafferongames.com/post/integration_basics/
    -   Chris Wilson's "Platformer Physics" talk/article (sometimes cited as "Higher-Order Character Physics"): A great explanation of platformer physics.
        -   *Link*: https://www.youtube.com/watch?v=hG9SzQxaCm8 (video)
        -   *Link*: http://www.somethinghitme.com/2013/11/11/simple-2d-platformer-physics-part-1/ (related blog post)

-   **Collision Detection:**
    -   MDN Web Docs - 2D collision detection: A simple guide with examples.
        -   *Link*: https://developer.mozilla.org/en-US/docs/Games/Techniques/2D_collision_detection
    -   "N" tutorial on collision detection for games:
        -   *Link*: https://www.metanetsoftware.com/technique/tutorial_a.html

-   **Game Development Patterns:**
    -   Game Programming Patterns by Robert Nystrom: An excellent book covering many patterns used in game development. The online version is free.
        -   *Link*: https://gameprogrammingpatterns.com/
