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

All components are defined in the `components/` directory. Each component is a simple struct that holds data. 

#### Pointer Wrapper Pattern (`ObjectData`)
A critical pattern used in this project is the **Pointer Wrapper** for external library objects. For example, `resolv.Object` is stored as:

```go
type ObjectData struct {
	*resolv.Object
}
var Object = donburi.NewComponentType[ObjectData]()
```

**Why?** Donburi stores component data in contiguous slices. When these slices grow, they reallocate, and the memory addresses of the components change. Libraries like `resolv` maintain their own pointers to these objects. By storing a pointer *inside* the component (the wrapper), we ensure that the address held by the library remains valid even if the component itself moves in memory.

#### State Management (`StateID`)
Characters use a type-safe `StateID` enum (defined in `config/states.go`) instead of strings:

```go
type StateData struct {
	CurrentState config.StateID
	StateTimer   int
}
```

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

Factories, located in the `systems/factory/` directory, are responsible for creating and initializing complex entities. They use archetypes to spawn new entities and then set their initial component values.

**Key Rule: Factory Ownership**
A factory function should "own" the entire creation process, including the creation of external physics objects. It should take simple parameters (position, size) rather than pre-created objects.

```go
// GOOD
func CreateWall(ecs *ecs.ECS, x, y, w, h float64) *donburi.Entry {
    // Factory creates the resolv object
    obj := resolv.NewObject(x, y, w, h, tags.ResolvSolid)
    // ...
}

// BAD
func CreateWall(ecs *ecs.ECS, obj *resolv.Object) *donburi.Entry {
    // Factory relies on caller to create object correctly
}
```

This approach encapsulates the complexity of entity creation and ensures that all entities are properly initialized with correct tags and data linkages.

### Game Initialization

The main game scene, `scenes/world.go`, is where everything comes together. In the `configure` method, we:

1.  Create a new `ecs.ECS` instance.
2.  Add all the systems and renderers to the ECS.
3.  Use factories to create the initial game entities (level, player, camera, etc.).

## Configuration Management

We use a **Centralized Configuration** pattern. All game tuning values (speed, damage, health, dimensions) are stored in `config/config.go`.

*   **Initialization**: All configuration structs are populated in the `init()` function of `config/config.go`.
*   **Usage**: Systems and factories read directly from `config.Player`, `config.Combat`, etc.
*   **Avoid**: Do not hardcode values in systems (e.g., `damage := 10`). Do not initialize config values in distributed `init()` functions across different packages.

## Physics & Resolv Integration

We use `solarlune/resolv` for collision detection. To ensure high performance and correct ECS integration:

1.  **Reverse Lookup (`Data` Field)**: When creating a `resolv.Object`, **ALWAYS** set its `Data` field to point to the `donburi.Entry`.
    ```go
    obj := resolv.NewObject(x, y, w, h)
    obj.Data = entry // Critical for O(1) reverse lookup
    ```
2.  **Optimized Collision**: Use `obj.Check(...)` which utilizes the `resolv.Space` spatial partition (O(log N)). Avoid iterating all entities and checking `Shape.Intersection` (O(N)).
3.  **Tag Constants**: Use defined constants in `tags/tags.go` (e.g., `tags.ResolvEnemy`) instead of string literals like "Enemy" or "solid".

## Best Practices Checklist

1.  **State Safety**: Always add new character or game states to `config/states.go` as `StateID` constants.
2.  **Zero Allocation**: Avoid creating new objects (like `ebiten.DrawImageOptions` or `color.RGBA`) inside `Draw` or `Update` loops. Reuse package-level variables.
3.  **Component Caching**: When iterating over entities in a system, if you need to access multiple components, call `Get` once at the start of the loop body.
4.  **Pointer Wrappers**: Use `ObjectData` to store pointers to external objects (like `resolv.Object`) to prevent memory invalidation.
5.  **Factory Integrity**: Factories must fully initialize the entity, including setting `obj.Data` and correct tags.
6.  **Config Centralization**: All magic numbers belong in `config/config.go`.