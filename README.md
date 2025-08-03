# Doomerang

Doomerang is a 2D fighting/platforming game built using the Ebiten game engine.
                                                                                
## Directory Structure                                                          
                                                                                
The project follows a structure common in ECS-based game development:           
                                                                                
-   `main.go`: The application entry point. It initializes Ebiten and starts the
main game scene.                                                                
                                                                                
-   `/scenes`: Manages different game states or screens. `platformer.go`        
contains the primary gameplay scene, which sets up the ECS world and its        
systems.                                                                        
                                                                                
-   `/components`: Defines the data-only components in the ECS. For example,    
`PlayerData` holds state for the player character, and `AnimationData` holds    
animation state.                                                                
                                                                                
-   `/systems`: Contains the logic that operates on entities with specific      
components.                                                                     
    -   `player.go`: Handles player input, physics, and animation state changes.
    -   `enemy.go`: Handles enemy animation state, physics
    -   `camera.go`: Controls camera movement to follow the player.             
    -   `level.go`, `objects.go`: Manages level rendering and object updates.   
    -   `debug.go`: Renders debug information when enabled.                     
                                                                                
-   `/archetypes`: Defines templates for creating entities. An archetype is a   
pre-configured set of components (e.g., a "Player" archetype has position,      
physics, and player-specific components).                                       
                                                                                
-   `/factory`: Provides functions to create game entities (like the player,    
platforms, enemies) using the defined archetypes. This encapsulates the         
complexity of entity creation.                                                  
                                                                                
-   `/assets`: Contains all game assets.                                        
    -   `/levels`: Tiled (`.tmx`) map files define the level layouts.           
    -   `/images`: Spritesheets and other images.                               
    -   `/fonts`: Font files.                                                   
                                                                                
-   `/resolv`: A small wrapper package to integrate the `resolv` collision      
library with the Donburi ECS. It provides helper functions for managing physics 
objects within the ECS world.                                                   
                                                                                
-   `/config`: Holds global configuration values like screen dimensions.

## How to play

Arrow key to move, X to jump.

### DEBUG Usage
To run with debug output:
`DEBUG_COLLISION=1 go run main.go`
.
To run without debug output (normal mode):
`go run main.go`

When DEBUG_COLLISION is set, you'll see console output like:
```
Horizontal collision detected! dx=2.50, player pos: (150.00, 200.00)
  Solid 0: pos=(0.67, 65.00), size=(31.00, 255.67)
  Solid 1: pos=(0.00, 321.00), size=(640.00, 47.00)
```

This will help you debug collision issues in the future by showing:
The attempted movement distance (dx)
The player's current position
All solid objects that are being detected in the collision check
Their positions and sizes
The debug output only appears when there's a horizontal collision detection, so it won't spam the console during normal gameplay.