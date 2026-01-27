## Goal
Add a complete audio system with looping background music (menu + per-level) and sound effects for gameplay actions.

## Design

### Music System

**Source Location:** `assets/art-archive/asset-pack-raw/Audio/music/`

**Available Tracks:**
1. `DavidKBD - InterstellarPack - 01 - Interstellar.ogg`
2. `DavidKBD - InterstellarPack - 02 - Plasma Storm.ogg`
3. `DavidKBD - InterstellarPack - 03 - Temple of Madness.ogg`
4. `DavidKBD - InterstellarPack - 04 - Horsehead Nebula.ogg`
5. `DavidKBD - InterstellarPack - 05 - Forgotten Station.ogg`
6. `DavidKBD - InterstellarPack - 06 - Hope on the Horizon.ogg`
7. `DavidKBD - InterstellarPack - 07 - Electric Firework.ogg`
8. `DavidKBD - InterstellarPack - 08 - Synth Kobra.ogg`
9. `DavidKBD - InterstellarPack - 09 - Spiral of Plasma.ogg`

**Menu Music:**
- Looping track that plays on the main menu
- Suggested: `Hope on the Horizon` or `Interstellar` (more mellow)
- Fades out when starting a level

**Level Music:**
- Each level has its own music that loops
- Convention: `assets/levels/level01/music/` contains the track
- One song per level, loops when finished
- Suggested for Level 1: `Plasma Storm` or `Electric Firework` (high energy)
- Asset pack: https://davidkbd.itch.io/interstellar-edm-metal-music-pack

### Sound Effects

**Asset Packs:**
- https://obsydianx.itch.io/interface-sfx-pack-1 (UI sounds)
- https://heltonyan.itch.io/pixelcombat (combat sounds)

**Source Location:** `assets/art-archive/asset-pack-raw/Audio/sfx/Helton Yan's Pixel Combat - Single Files/`

**SFX Mapping (suggested files):**

| Sound | Suggested Asset |
|-------|-----------------|
| **Combat - Punch/Kick** | `FGHTImpt_MELEE-Gut Punch`, `FGHTImpt_MELEE-Gut Kick`, `FGHTImpt_MELEE-Crunch Kick` |
| **Combat - Hit Impact** | `FGHTImpt_HIT-Strong Punch`, `FGHTImpt_HIT-Smack`, `DSGNMisc_HIT-Sweep Hit` |
| **Combat - Enemy Death** | `DSGNImpt_EXPLOSION-Crunchy Burst`, `DSGNMisc_SKILL IMPACT-Dramatic Finish` |
| **Movement - Jump** | `DSGNMisc_MOVEMENT-Retro Jump`, `DSGNMisc_MOVEMENT-Jump Sparkle` |
| **Movement - Land** | `FEETMisc_STEP-Hard Step`, `FEETMisc_STEP-Boots on Concrete` |
| **Movement - Slide** | `SWSH_MOVEMENT-Sand Swipe`, `DSGNMisc_MOVEMENT-Tire Screech` |
| **Boomerang - Quick Throw** | `DSGNMisc_PROJECTILE-High Whoosh`, `WHSH_MOVEMENT-Simple Whoosh` |
| **Boomerang - Charged Throw** | `DSGNTonl_SKILL RELEASE-Laser Whoosh 1`, `DSGNMisc_SKILL RELEASE-Flying Blades` |
| **Boomerang - Flight Loop** | `WHSH_MOVEMENT-Wind Sweep Swish`, `DSGNMisc_MOVEMENT-Zappy Shimmer` |
| **Boomerang - Catch** | `DSGNTonl_USABLE-Coin Toss`, `UIClick_INTERFACE-Positive Click` |
| **Boomerang - Enemy Hit** | `DSGNImpt_EXPLOSION-Electric Hit`, `DSGNMisc_SKILL IMPACT-Critical Strike` |
| **UI - Navigate** | `UIClick_INTERFACE-Metallic Click`, `DSGNTonl_INTERFACE-Tonal Click` |
| **UI - Select** | `UIClick_INTERFACE-Strong Click 1`, `UIMisc_INTERFACE-Zap Select` |
| **UI - Denied** | `UIMisc_INTERFACE-Denied` |
| **Player - Damage** | `DSGNMisc_HIT-Noisy Hit`, `FGHTImpt_HIT-Strong Smack` |
| **Player - Death** | `DSGNImpt_EXPLOSION-Forced Shutdown`, `DSGNSynth_BUFF-Failed Buff` |

### Volume Control

- **Music Volume:** Separate slider (0, 25%, 50%, 75%, 100%)
- **SFX Volume:** Separate slider (0, 25%, 50%, 75%, 100%)
- Settings persist (ties into Feature 10: Settings Screen)

## Implementation Tasks

### 4.1 Audio System Foundation
- [ ] Add audio manager/system for playing sounds
- [ ] Implement music playback with looping
- [ ] Implement SFX playback (one-shot sounds)
- [ ] Add volume control (music + SFX separate)

### 4.2 Music Integration
- [ ] Add menu music that plays on main menu
- [ ] Load level music from `assets/levels/[level]/music/`
- [ ] Transition music when entering/exiting levels
- [ ] Ensure music loops seamlessly

### 4.3 Sound Effects Integration
- [ ] Import and organize SFX assets
- [ ] Add combat sounds (punch, kick, hit, death)
- [ ] Add movement sounds (jump, land, slide)
- [ ] Add boomerang sounds (throw, flight, catch, impact)
- [ ] Add UI sounds (menu navigation, select)

### 4.4 Settings Integration
- [ ] Store volume settings (prepare for Settings Screen)
- [ ] Add mute toggle support

## Files to Modify/Create
- `systems/audio.go` - New audio system
- `components/audio.go` - Audio component if needed
- `config/config.go` - Volume defaults, audio settings
- `assets/assets.go` - Audio asset loading
- `assets/music/` - Music files
- `assets/sfx/` - Sound effect files
- `scenes/menu.go` - Menu music
- `scenes/world.go` - Level music

## Asset Organization
```
assets/
├── music/
│   └── menu/
│       └── menu-loop.ogg
├── levels/
│   └── level01/
│       └── music/
│           └── track.ogg
└── sfx/
    ├── combat/
    ├── movement/
    ├── boomerang/
    └── ui/
```
