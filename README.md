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
                                                                                
-   `/resolv`: A small wrapper package to integrate the `resolv` collision      
library with the Donburi ECS. It provides helper functions for managing physics 
objects within the ECS world.                                                   
                                                                                
-   `/config`: Holds global configuration values like screen dimensions and types of valid global states.
