# Doomerang

Doomerang is a 2D fighting/platforming game built using the Ebiten game engine.
                                                                                
## Directory Structure                                                          
                                                                                
The project follows a structure common in ECS-based game development:           
                                                                                
-   `main.go`: The application entry point. It initializes Ebiten and starts the
main game scene.                                                                
                                                                                
-   `/scenes`: Manages different game states or screens. `world.go`
contains the primary gameplay scene, which sets up the ECS world and its
systems.
                                                                                
-   `/components`: Defines the data-only components in the ECS. For example,    
`PlayerData` holds state for the player character, and `AnimationData` holds    
animation state.                                                                
                                                                                
-   `/systems`: Contains the logic that operates on entities with specific      
components.
    - `/factory`: Provides functions to create game entities (like the player,    
platforms, enemies) using the defined archetypes. This encapsulates the         
complexity of entity creation.                                                                      
    -   `player.go`: Handles player input and state changes.
    -   `enemy.go`: Handles enemy AI and state changes.
    -   `physics.go`: Handles physics calculations for all entities.
    -   `collision.go`: Handles collision detection and resolution for all entities.
    -   `render.go`: Handles rendering for all animated entities.
    -   `camera.go`: Controls camera movement to follow the player.
    -   `level.go`, `objects.go`: Manages level rendering and object updates.
    -   `debug.go`: Renders debug information when enabled.
                                                                                
-   `/archetypes`: Defines templates for creating entities. An archetype is a   
pre-configured set of components (e.g., a "Player" archetype has position,      
physics, and player-specific components).                                                                                                                       
                                                                              
-   `/assets`: Contains all game assets.                                        
    -   `/levels`: Tiled (`.tmx`) map files define the level layouts.           
    -   `/images`: Spritesheets and other images.                               
    -   `/fonts`: Font files.                                                   
    -   `assets.go`: Handles asset loading with built-in **caching** to prevent redundant decoding.

-   `/config`: Holds global configuration values, constants, and the type-safe `StateID` system.

## Performance & Architecture

The project has been optimized for high performance and stability:

-   **Asset Caching**: Images are decoded once and cached in memory. Subsequent requests for the same sprite sheet return the cached pointer, drastically reducing CPU and memory overhead during level loads and entity creation.
-   **Zero-Allocation Rendering**: The render systems reuse global `DrawImageOptions` and pre-defined `color.Color` variables to eliminate per-frame heap allocations, preventing GC stuttering.
-   **Type-Safe States**: All character and game states use the `StateID` enum (defined in `config/states.go`). This prevents typo-related bugs and improves comparison speed.
-   **Memory Safety**: Physics objects (`resolv.Object`) are stored via a pointer wrapper (`ObjectData`) in ECS components. This ensures that the `resolv.Space` always has valid pointers even when Donburi reallocates component storage.
-   **O(1) Hitbox Lookup**: Entities maintain a direct reference to their active hitbox, eliminating O(N) searches in hot combat loops.

## Tags

In Donburi ECS, tags are special components used to label and identify entities without attaching complex data. They are defined in `tags/tags.go`.

### Why use tags?
- **Lightweight Identification**: Tags act as flags (e.g., `Player`, `Enemy`, `Platform`) to easily categorize entities.
- **Filtering Queries**: They allow systems to efficiently query for specific groups of entities. For example, the `render` system might query all entities with a `Player` tag to apply player-specific rendering logic.

### Usage
- **Defining Tags**: Tags are defined as exported variables in `tags/tags.go` using `donburi.NewTag().SetName("TagName")`.
- **Adding to Entities**: Tags are added to entities during creation, typically within Archetypes (see `archetypes/archetypes.go`).
- **Querying**: Systems can use tags to iterate over specific entities:
  ```go
  // Example: Iterate over all entities with the Player tag
  tags.Player.Each(ecs.World, func(e *donburi.Entry) {
      // Logic for player entity
  })
  
  // Example: Check if a specific entity has a tag
  if e.HasComponent(tags.Enemy) {
      // Logic for enemy entity
  }
  ```
