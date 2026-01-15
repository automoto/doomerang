# Level Update Guide: Retro Cyberpunk Theme

## 1. Asset Suggestions
Based on the `asset-pack-raw` content, here are the recommended assets for a "Retro Cyberpunk" aesthetic:

### Tiles
*   **Primary Ground/Platforms**: `Tile Set/Assets/Tile Set/Street Tiles/Stylish Black`. This set provides a sleek, dark urban look suitable for a cyberpunk street level.
    *   Use `Center-Solid.png` or `Center.png` for fillers.
    *   Use the `Curve.*` and `Edge.*` files for organic or sharp corners.
*   **Industrial/Indoor Sections**: `Tile Set/Assets/Tile Set/Metal Tiles` (check subfolders like `Rusty` or `Clean` if available) or `Interior Tiles` for variety.
*   **Platforms**: `Tile Set/Assets/Tile Set/Street Tiles/Dirty Street Blue` can be used for contrast or specific zones (e.g., lower, "dirtier" levels).

### Backgrounds
*   **Main Background**: `Tile Set/Assets/Building Set/Loopable Background/Background-Night.png`. This is the quintessential dark, atmospheric cyberpunk backdrop.
*   **Parallax Layers**: 
    *   `City-Block-Far.png` (slow moving)
    *   `City-Block.png` (medium speed)
    *   `Moon-Night.png` (static or very slow)
    *   Use the "Blurred" versions (e.g., `City-Block-Far-Blurred-2.3%.png`) for depth of field effects if the engine supports it.

### Props & Decor
*   **Atmosphere**: `Tile Set/Assets/Enviroment and Props/Lamps/Lamp Street (Side)` and `Lamp LED` for lighting.
*   **Cyberpunk Elements**: `Street Signs`, `Advertising Panels & Signs` (add neon flavor), `Cyberpunk Coin Animated` (pickups).
*   **Obstacles**: `Traps/Spikes`, `Traps/Mines`, `Enviroment and Props/Road Block`.

## 2. Execution Steps (Using Tiled)

### A. Importing the New Tileset
1.  **Open Tiled** and load your level (`.tmx` file).
2.  **New Tileset**: Click the "New Tileset" button (or `File > New > New Tileset`).
    *   **Type**: "Collection of Images" is best here because the source assets are individual images (`Center.png`, `Curve.TopLeft.png`, etc.) rather than a single spritesheet.
    *   **Source**: Navigate to `assets/art-archive/asset-pack-raw/Tile Set/Assets/Tile Set/Street Tiles/Stylish Black`.
    *   **Name**: `Cyberpunk_Street_Black`.
    *   Save the `.tsx` file in your `assets/levels/tilesets` or similar folder (create one if needed).
    *   *Tip*: You can drag and drop all the PNG files from your file browser into the Tileset Editor to add them quickly.

### B. Adding the Background
1.  **Create Image Layer**: In the "Layers" panel, right-click and select `New > Image Layer`.
2.  **Name it**: `Background`. Move it to the bottom of the list.
3.  **Set Image**: In the "Properties" panel for this layer, click the `Image` field and browse to `assets/art-archive/asset-pack-raw/Tile Set/Assets/Building Set/Loopable Background/Background-Night.png`.
4.  **Parallax**: If you want parallax, create additional Image Layers (e.g., `Background_City_Far`) using the city block images. In your game engine, you will likely need to write code to scroll these based on camera movement, but Tiled lets you position the initial state.

### C. Updating the Level (Tile Replacement)
*Option 1: Manual Redraw (Recommended for major style changes)*
1.  Select the `Cyberpunk_Street_Black` tileset.
2.  Select the "Stamp Brush" (B).
3.  Paint over the existing geometry. Since the shapes (curves, edges) might differ from your placeholder tiles, manual painting ensures the best look.

*Option 2: Bucket Fill Replace (Faster for simple blocks)*
1.  Select the old tile you want to replace in your map (right-click it in the map view to pick it).
2.  Select the "Bucket Fill Tool" (F).
3.  Select the *new* tile from your `Cyberpunk_Street_Black` tileset.
4.  Hold `Shift` and click the old tile in the map. This replaces *all* instances of that old tile with the new one. (Note: verify this shortcut in your specific Tiled version, or use the "Mass Replace" tool if installed).

### D. Adding Props
1.  **New Object Layer**: If your props need logic (collisions, interactions), use an "Object Layer". If they are just visual, a "Tile Layer" (e.g., `Foreground_Props`) is fine.
2.  **Import Props**: Create another "Collection of Images" tileset for `Props`. Drag in images from `assets/art-archive/asset-pack-raw/Tile Set/Assets/Enviroment and Props`.
3.  **Place**: Drag and drop props (Lamps, Signs) into the scene.

## 3. Cleanup
1.  **Remove Old Tilesets**: Once replaced, click the small "trash can" icon in the Tilesets panel to remove the placeholder tileset dependencies from your map file.
2.  **Save**: Save your `.tmx` map.
