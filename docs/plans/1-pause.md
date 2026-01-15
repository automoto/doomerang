## Goal
Implement a pause screen overlay that stops the game loop and offers three options: "Resume", "Settings", and "Exit".

## Design
- **Toggle**: Pressing `ESC` or `P` toggles the pause state.
- **Visuals**: A semi-transparent black overlay covers the game screen. The menu options are displayed in the center.
- **Navigation**: 
    - `UP` / `DOWN` keys to cycle through options.
    - `ENTER` to select an option.
- **Behavior**:
    - **Resume**: Unpauses the game.
    - **Settings**: Logs "Settings clicked" (placeholder for future implementation).
    - **Exit**: Closes the application.