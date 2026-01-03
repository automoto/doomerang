# Doomerang

Doomerang is a 2D fighting/platforming game built using the Ebiten game engine with an ECS (Entity-Component-System) architecture.

## Quick Start

**Prerequisites:** Go 1.23+, golangci-lint (for linting)

```bash
make build       # Compile the game
make run         # Run the game
make lint        # Check code quality
make basic-test  # Run integration test
```

## Directory Structure

- `main.go` - Application entry point; initializes Ebiten and starts the main game scene
- `/scenes` - Game states/screens; `world.go` sets up the ECS world and systems
- `/components` - Data-only ECS components (e.g., `PlayerData`, `AnimationData`)
- `/systems` - Game logic operating on entities with specific components
  - `/factory` - Entity creation functions using defined archetypes
- `/archetypes` - Entity templates (pre-configured component bundles)
- `/assets` - Game assets with built-in caching
  - `/levels` - Tiled (`.tmx`) map files
  - `/images` - Spritesheets
  - `/fonts` - Font files
- `/config` - Configuration values, `StateID` system, and input bindings
- `/tags` - Donburi tags for entity categorization and filtering

## Performance & Architecture

The project has been optimized for high performance and stability:

- **Asset Caching**: Images are decoded once and cached in memory. Subsequent requests for the same sprite sheet return the cached pointer, drastically reducing CPU and memory overhead during level loads and entity creation.
- **Zero-Allocation Rendering**: The render systems reuse global `DrawImageOptions` and pre-defined `color.Color` variables to eliminate per-frame heap allocations, preventing GC stuttering.
- **Type-Safe States**: All character and game states use the `StateID` enum (defined in `config/states.go`). This prevents typo-related bugs and improves comparison speed.
- **Memory Safety**: Physics objects (`resolv.Object`) are stored via a pointer wrapper (`ObjectData`) in ECS components. This ensures that the `resolv.Space` always has valid pointers even when Donburi reallocates component storage.
- **O(1) Hitbox Lookup**: Entities maintain a direct reference to their active hitbox, eliminating O(N) searches in hot combat loops.
- **ECS Optimization**: Redundant ECS operations are minimized by caching component checks in hot loops, preventing state tag thrashing via change detection, and caching configuration pointers to avoid expensive map lookups.
- **Input Abstraction**: Raw input polling is decoupled from game logic via an `InputData` component. The `UpdateInput` system maps keys/buttons to logical actions (`ActionJump`, `ActionAttack`, etc.), allowing easy remapping and multi-input support.

Tags are used for lightweight entity identification and filtering. See [ECS documentation](docs/ECS_AND_DONBURI.md) for details.

## Documentation

- [ECS & Donburi Architecture](docs/ECS_AND_DONBURI.md) - Comprehensive guide to the ECS pattern and donburi usage
- [Game Math & Physics](docs/game-math-physics.md) - Physics implementation details
- [Platformer Physics Tutorial](docs/TUTORIAL.md) - Educational guide to game physics concepts
- [Enemy AI](docs/enemy_ai.md) - Enemy behavior and state machine documentation
