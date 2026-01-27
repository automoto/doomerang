## Goal
Implement a main menu screen that serves as the game's entry point, displaying the title and navigation options.

## Design

**Layout:**
- Dark/minimal background (similar to pause overlay)
- "Doomerang" title at top center - larger font, distinct color for emphasis
- Menu options centered below title

**Menu Options:**
1. **Start** - Begins a new game
2. **Continue** - Placeholder: logs "Continue clicked" (implemented with Save Games feature)
3. **Level Select** - Placeholder: logs "Level Select clicked" (implemented with Level system)
4. **Settings** - Placeholder: logs "Settings clicked"
5. **Exit** - Closes the application

**Navigation:**
- `UP` / `DOWN` keys to cycle through options
- `ENTER` to select an option
- Visual indicator for currently selected option

**Behavior:**
- Menu displays immediately on game launch (no title screen)
- Selecting "Start" transitions to gameplay
- Placeholder options log their selection for now

## Linked Tasks
When implementing related features, wire up the placeholder menu options:
- Feature 9 (Save Games): Implement "Continue" functionality
- Feature 7/8 (Levels): Implement "Level Select" functionality
- Future: Implement "Settings" screen
