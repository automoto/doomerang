# Doomerang

Doomerang is a fast-paced 2D action-platformer built with **Go** and the [Ebitengine](https://ebitengine.org/) game engine.

<img width="534" height="321" alt="Screenshot 2026-01-28 at 3 33 02 AM" src="https://github.com/user-attachments/assets/94d9e5d8-c741-46b0-98f2-66abde1cfa26" />

## Quick Start

### Prerequisites
- [Go 1.21+](https://go.dev/dl/)
- OS-specific dependencies (see [Ebitengine installation](https://ebitengine.org/en/documents/install.html))

### Run
```bash
make run
```

## Architecture

The project follows a standard ECS (Entity Component System) pattern:

- `/components`: Data-only structures (e.g., `PlayerData`, `AnimationData`).
- `/systems`: Logic operating on entities (e.g., `UpdatePhysics`, `Render`).
- `/factory`: Entity creation functions using archetypes.
- `/scenes`: Game state management (Menu, World).
- `/assets`: Tiled maps, spritesheets, and audio.
- `/config`: Global constants, states, and input bindings.

For a deeper dive into the ECS implementation, see [ECS_AND_DONBURI.md](docs/ECS_AND_DONBURI.md).

### Technical Highlights

- Zero-Allocation Rendering: Reuses drawing options to eliminate per-frame GC pressure.
- Asset Caching: Centralized system prevents redundant asset decoding.
- Memory Safety: Uses pointer wrappers (`ObjectData`) for stable physics references during ECS reallocation.
- Type-Safe States: Centralized `StateID` enum prevents typo-related bugs.
- Input Abstraction: Decouples raw input from logic via action mapping.

> **macOS Note**: The Makefile uses `CGO_CFLAGS="-w"` to suppress Metal backend warnings.

### Assets

Assets are a mix of assets I sourced and some that I created. The following asset packs are used in making this game:
- [Fire Effects](https://stealthix.itch.io/animated-shots)
- [Other Effects](https://stealthix.itch.io/animated-effects)
- [Character](https://deadrevolver.itch.io/pixel-prototype-player-sprites)
- [Tiles](https://overcrafted.itch.io/cyberpunk-platformer-tileset)
- [Music](https://davidkbd.itch.io/interstellar-edm-metal-music-pack)
- [Sound Effects](https://heltonyan.itch.io/pixelcombat)
