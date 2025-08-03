# 2D Platformer Math and Physics Tutorial

This document provides a concise overview of the essential math and physics concepts required to build a 2D platformer. It includes explanations, code examples, and links to further resources.

## 1. Core Concepts

### Vectors

In a 2D game, we use vectors to represent quantities that have both a magnitude (amount) and a direction. The most common uses are for position, velocity, and acceleration. A 2D vector is typically represented by two components: `x` and `y`.

-   **Position**: Where an object is in the game world (e.g., `(100, 50)`).
-   **Velocity**: The rate of change of an object's position. It represents speed and direction (e.g., `(5, 0)` means moving right at 5 pixels per frame).
-   **Acceleration**: The rate of change of an object's velocity (e.g., gravity is a downward acceleration, `(0, 0.5)`).

**Further Learning:**
-   **Video**: [Math for Game Developers - Vectors by Freya Holmér](https://www.youtube.com/watch?v=MOYiVLEnhrw) - An excellent visual introduction to vectors.
-   **Book**: "Game Programming Patterns" by Robert Nystrom has a chapter on [Vectors](http://gameprogrammingpatterns.com/vectors.html).

### The Game Loop and Delta Time

A game runs in a loop. In each iteration of the loop, it processes input, updates the game state, and draws everything to the screen.

```
while (game is running) {
    processInput();
    update();
    render();
}
```

The `update()` function is where most of our physics calculations happen. To make movement independent of the computer's speed (framerate), we use a concept called **delta time** (`dt`). `dt` is the time elapsed since the last frame.

All physics calculations should be multiplied by `dt` to ensure the game runs consistently across different hardware.

`position += velocity * dt;`
`velocity += acceleration * dt;`

**Further Learning:**
-   **Article**: [Fix Your Timestep! by Glenn Fiedler](https://gafferongames.com/post/fix_your_timestep/) - A classic, in-depth article about game loops and timing.
-   **Book**: "Game Programming Patterns" by Robert Nystrom has a great chapter on the [Game Loop](http://gameprogrammingpatterns.com/game-loop.html).

### Kinematics: The Equations of Motion

Kinematics is the study of motion. For a simple platformer, we only need a couple of basic equations, which are applied every frame in the `update` loop. These are known as Euler integration, a simple way to approximate continuous motion.

1.  **Update Velocity**: `velocity = velocity + (acceleration * dt)`
2.  **Update Position**: `position = position + (velocity * dt)`

In our game code (`systems/player.go`), this looks like:
- Gravity is a constant downward acceleration applied to the player's `SpeedY`.
- Player input creates acceleration that modifies `SpeedX`.
- The player's object position is then updated using `SpeedX` and `SpeedY`.

**Further Learning:**
-   **Tutorial**: [Khan Academy on Kinematic Formulas](https://www.khanacademy.org/science/physics/one-dimensional-motion/kinematic-formulas/a/what-are-the-kinematic-formulas) - Provides the physics background.

## 2. Player Movement Physics

### Horizontal Movement

-   **Acceleration**: When the player presses a move key (left/right), we apply an acceleration in that direction. This gradually increases the player's velocity.
-   **Friction**: When no key is pressed, we need to slow the player down. This is done by applying an opposing acceleration (friction) until the velocity is zero.
-   **Max Speed**: To prevent the player from accelerating infinitely, we cap the velocity at a maximum value.

In `systems/player.go`, these are controlled by `playerAccel`, `playerFriction`, and `playerMaxSpeed`.

### Jumping

-   **Gravity**: A constant downward acceleration applied to the player every frame. This is what makes the player fall. `playerGravity` in the code.
-   **Jump Impulse**: When the jump button is pressed, we give the player an instant, large upward velocity. This is a "jump impulse". It's enough to overcome gravity for a short time. `playerJumpSpd` in the code.

**Variable Jump Height**
A common feature is to allow the player to jump higher the longer they hold the jump button. A simple way to implement this is:
- If the player releases the jump button while they are still moving upwards, reduce their upward velocity (e.g., set `player.SpeedY` to a smaller value if it's negative). This will cut the jump short.

**Further Learning:**
-   **Article**: [N/N+ Platformer Physics Tutorial](https://www.metanetsoftware.com/technique/tutorial-a-platformer-physics/) - A classic and detailed guide to platformer physics from the creators of N.
-   **Video**: [The "Feel" of Jumping in Platformers by Game Maker's Toolkit](https://www.youtube.com/watch?v=c3iEl5AwUF8) - Discusses what makes a jump feel good, including variable jump height.

### Wall Sliding and Wall Jumping

-   **Wall Sliding**: If the player is in the air and pressing against a wall, their downward speed due to gravity is reduced. This is implemented by checking for a wall collision and then setting a slower terminal velocity.
-   **Wall Jumping**: If the player is wall-sliding and presses the jump button, they perform a special jump that pushes them up and away from the wall. This is an impulse applied both vertically and horizontally away from the wall.

## 3. Collision Detection and Resolution

Collision detection is figuring out if two shapes are overlapping. Collision resolution is what you do about it. Modern game engines and physics libraries (like `resolv` used in this project) often handle most of this for you.

### Axis-Aligned Bounding Boxes (AABB)

The simplest and most common collision shape is the Axis-Aligned Bounding Box (AABB). It's a rectangle that is not rotated. Checking for an overlap between two AABBs is very fast and simple.

To check if two AABBs (A and B) are colliding:
`A.x < B.x + B.width && A.x + A.width > B.x && A.y < B.y + B.height && A.y + A.height > B.y`

### Collision Resolution

When a moving object (like the player) collides with a static object (like a wall), we need to stop it from passing through. The simplest way is:
1.  Check for a collision along the X-axis first. If there is one, move the player so their side is just touching the wall and set their X velocity to 0.
2.  Then, check for a collision along the Y-axis. If there is one (e.g., landing on a floor), move the player so their bottom is just touching the floor and set their Y velocity to 0. This also signals that the player is `OnGround`.

Separating the axes like this prevents buggy behavior, like getting stuck on corners.

**Further Learning:**
-   **Article**: [MDN 2D collision detection](https://developer.mozilla.org/en-US/docs/Games/Techniques/2D_collision_detection) - A great tutorial on various 2D collision techniques.
-   **Tutorial**: [The Guide to Implementing 2D Platformers](https://www.higherorderfun.com/blog/2021/04/03/the-guide-to-implementing-2d-platformers/) - A comprehensive guide to tile-based collision.
-   **Citation**: This project uses the [resolv](https://github.com/solarlune/resolv) library, which handles AABB-based collision detection and resolution. Its documentation is a good resource for understanding how it works internally.
