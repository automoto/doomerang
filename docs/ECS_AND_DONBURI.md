# Entity Component System (ECS) and the `donburi` Library

This document provides a concise explanation of the Entity Component System (ECS) architecture and how it is implemented in this project using the `donburi` library. The goal is to establish a clear and consistent approach to game development that can be easily understood and followed by other developers.

## What is ECS?

ECS is a software architectural pattern primarily used in game development. It follows the principle of "composition over inheritance," which helps to create flexible and efficient game logic. The architecture is built on three core concepts:

*   **Entities**: An entity is a unique identifier for a game object. It's essentially a number that has no data or behavior associated with it. Think of it as a key in a database.
*   **Components**: Components are plain data structures that hold the state of an entity. They do not contain any logic. For example, a `PositionComponent` might store an entity's `(x, y)` coordinates, and a `VelocityComponent` would store its speed and direction.
*   **Systems**: Systems are responsible for implementing the game's logic. They operate on entities that have a specific set of components. For example, a `MovementSystem` would query the ECS world for all entities that have both a `PositionComponent` and a `VelocityComponent` and then update their positions based on their velocities.

## Why Use ECS?

The primary benefits of using an ECS architecture are:

*   **Flexibility**: Because entities are defined by their components, it's easy to create new types of game objects by simply mixing and matching components. This avoids the rigid class hierarchies that are common in object-oriented programming.
*   **Performance**: ECS is very cache-friendly. Since components are stored contiguously in memory, systems can iterate over them very efficiently. This leads to significant performance gains, especially in games with a large number of entities.
*   **Decoupling**: The separation of data (components) from logic (systems) leads to a highly decoupled codebase. This makes it easier to reason about the game's behavior, add new features, and reuse code.

## How We Use `donburi`

In this project, we use the `donburi` library to implement the ECS architecture. `donburi` provides a simple and efficient way to manage entities, components, and systems. Here's how we've organized our code:

### Components

All components are defined in the `components/` directory. Each component is a simple struct that holds data. For example, in `components/physics.go`, we have:

```go
package components

import (
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
)

type PhysicsData struct {
	SpeedX         float64
	SpeedY         float64
	AccelX         float64
	Gravity        float64
	Friction       float64
	MaxSpeed       float64
	OnGround       *resolv.Object
	WallSliding    *resolv.Object
	IgnorePlatform *resolv.Object
}

var Physics = donburi.NewComponentType[PhysicsData]()
```

Each component type is registered with `donburi` using `donburi.NewComponentType[T]()`, where `T` is the struct that defines the component's data.

### Systems

Systems contain the game's logic and are located in the `systems/` directory. A system is a function that takes an `*ecs.ECS` instance as an argument. For example, `systems/physics.go` defines the `UpdatePhysics` system:

```go
package systems

import (
	"github.com/automoto/doomerang/components"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func UpdatePhysics(ecs *ecs.ECS) {
	components.Physics.Each(ecs.World, func(e *donburi.Entry) {
		// ...
	})
}
```

Systems query the ECS world for entities that have the components they're interested in and then perform operations on that data.

### Archetypes

To simplify the creation of entities, we use archetypes, which are defined in `archetypes/archetypes.go`. An archetype is a pre-defined set of components that represents a type of entity. For example, the `Player` archetype is defined as:

```go
package archetypes

// ...

var (
    Player = newArchetype(
        tags.Player,
        components.Player,
        components.Object,
        components.Animation,
        components.Physics,
    )
)
```

This makes it easy to create a new player entity with all the necessary components.

### Factories

Factories, located in the `factory/` directory, are responsible for creating and initializing complex entities. They use archetypes to spawn new entities and then set their initial component values. For example, `factory/player.go` has a `CreatePlayer` function:

```go
package factory

// ...

func CreatePlayer(ecs *ecs.ECS) *donburi.Entry {
    player := archetypes.Player.Spawn(ecs)

    // ... initialize player components ...

    return player
}
```

This approach encapsulates the complexity of entity creation and ensures that all entities are properly initialized.

### Game Initialization

The main game scene, `scenes/world.go`, is where everything comes together. In the `configure` method, we:

1.  Create a new `ecs.ECS` instance.
2.  Add all the systems and renderers to the ECS.
3.  Use factories to create the initial game entities (level, player, camera, etc.).

```go
package scenes

// ...

func (ps *PlatformerScene) configure() {
    ecs := ecs.NewECS(donburi.NewWorld())

    // Add systems
    ecs.AddSystem(systems.UpdatePlayer)
    ecs.AddSystem(systems.UpdatePhysics)
    ecs.AddSystem(systems.UpdateCollisions)
    // ...

    // Add renderers
    ecs.AddRenderer(layers.Default, systems.DrawAnimated)
    // ...

    ps.ecs = ecs

    // Create entities
    factory.CreateLevel(ps.ecs)
    factory.CreatePlayer(ps.ecs)
    // ...
}
```

By following these conventions, we can maintain a clean, organized, and efficient codebase that is easy to extend and maintain.
