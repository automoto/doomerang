# TMX Syntax Guide for Doomerang

Reference for creating and editing `.tmx` (Tiled Map XML) files by hand. Covers the TMX spec as it applies to this project, our tileset GID mappings, and project-specific conventions.

Official spec: https://doc.mapeditor.org/en/stable/reference/tmx-map-format/

---

## 1. Document Structure

A `.tmx` file is XML. The root element is `<map>`. Children appear in this order:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<map ...attributes...>
 <properties>...</properties>          <!-- optional: map-level properties -->
 <tileset firstgid="1" source="..."/> <!-- one or more tileset references -->
 <imagelayer ...>...</imagelayer>       <!-- optional: background images -->
 <layer ...>...</layer>                <!-- one or more tile layers -->
 <objectgroup ...>...</objectgroup>    <!-- one or more object groups -->
</map>
```

Order matters: `<properties>` first (if present), then `<tileset>`, then layers and object groups in the order they should be drawn/processed.

---

## 2. `<map>` Element

```xml
<map version="1.10" tiledversion="1.11.2"
     orientation="orthogonal" renderorder="right-down"
     width="20" height="20"
     tilewidth="16" tileheight="16"
     infinite="0"
     nextlayerid="5" nextobjectid="10">
```

| Attribute | Required | Description |
|-----------|----------|-------------|
| `version` | yes | TMX format version. Use `"1.10"`. |
| `tiledversion` | no | Tiled editor version. Use `"1.11.2"`. |
| `orientation` | yes | Always `"orthogonal"` for this project. |
| `renderorder` | no | Always `"right-down"`. |
| `width` | yes | Map width **in tiles**. |
| `height` | yes | Map height **in tiles**. |
| `tilewidth` | yes | Tile width in pixels. Always `16`. |
| `tileheight` | yes | Tile height in pixels. Always `16`. |
| `infinite` | no | Always `"0"` (finite maps). |
| `nextlayerid` | no | Must be greater than any `id` used on layers/objectgroups. |
| `nextobjectid` | no | Must be greater than any `id` used on objects. |

**Pixel dimensions** = `width * tilewidth` by `height * tileheight`.

---

## 3. `<properties>` Element

`<properties>` can be a child of: `<map>`, `<layer>`, `<objectgroup>`, `<object>`, `<imagelayer>`, `<tileset>`, `<tile>`.

```xml
<properties>
 <property name="key" value="string_value"/>
 <property name="count" type="int" value="42"/>
 <property name="speed" type="float" value="3.5"/>
 <property name="visible" type="bool" value="true"/>
 <property name="color" type="color" value="#ff00ff00"/>
</properties>
```

| `type` value | Go accessor in `go-tiled` | Notes |
|-------------|---------------------------|-------|
| *(omitted)* | `GetString(name)` | Default type is `string`. |
| `string` | `GetString(name)` | Explicit string. |
| `int` | `GetInt(name)` | Integer. |
| `float` | `GetFloat(name)` | Float64. |
| `bool` | `GetBool(name)` | `"true"` or `"false"`. |
| `color` | `GetString(name)` | `#AARRGGBB` or `#RRGGBB` format. |
| `file` | `GetString(name)` | File path. |
| `object` | `GetInt(name)` | Object ID reference. |

**Common mistake:** Omitting `type` for non-string properties. `type="int"` is required for integers -- without it, `GetInt()` returns 0.

---

## 4. `<tileset>` Element

We always use **external tilesets** (a `.tsx` file):

```xml
<tileset firstgid="1" source="../levels/tilesets/cyberpunk-tiles.tsx"/>
```

| Attribute | Description |
|-----------|-------------|
| `firstgid` | The Global ID assigned to the first tile in this tileset. Always `1` for our single tileset. |
| `source` | Relative path from the `.tmx` file to the `.tsx` file. |

### GID Calculation

```
GID = firstgid + tile_id
```

Where `tile_id` is the `id` attribute in the `.tsx` file. With `firstgid="1"`:

| GID | Tile ID | Image | Visual Use |
|-----|---------|-------|------------|
| 0 | -- | *(empty)* | Empty/air tile |
| 2 | 1 | stylish-black-16/Ground.Top.png | Floor surface (black tileset) |
| 3 | 2 | stylish-black-16/Ground.Right.png | Right wall (black tileset) |
| 4 | 3 | stylish-black-16/Ground.Left.png | Left wall (black tileset) |
| **16** | 15 | **dirty-street-blue/Ground.Top.png** | **Floor surface** |
| **17** | 16 | **dirty-street-blue/Ground.Right.png** | **Right boundary wall** |
| **18** | 17 | **dirty-street-blue/Ground.Left.png** | **Left boundary wall** |
| 19 | 18 | dirty-street-blue/Ground.Bottom.png | Ceiling / bottom of platform |
| 20 | 19 | dirty-street-blue/Edge.TopRight.png | Slope 45 up-left |
| 21 | 20 | dirty-street-blue/Edge.TopLeft.png | Slope 45 up-right |
| 22 | 21 | dirty-street-blue/Edge.BottomRight.png | Corner piece |
| 23 | 22 | dirty-street-blue/Edge.BottomLeft.png | Corner piece |
| 24 | 23 | dirty-street-blue/Curve.TopRight.png | Rounded corner |
| 25 | 24 | dirty-street-blue/Curve.TopLeft.png | Rounded corner |
| 26 | 25 | dirty-street-blue/Curve.BottomRight.png | Rounded corner |
| 27 | 26 | dirty-street-blue/Curve.BottomLeft.png | Rounded corner |
| **28** | 27 | **dirty-street-blue/Center.png** | **Fill / solid interior** |
| 29 | 28 | dirty-street-blue/Center-Drain.png | Fill variant |
| 68 | 67 | interior-16/Ground.Top.png | Floor (interior tileset) |

**Bold** entries are the most commonly used tiles.

### Quick Reference for Common Patterns

```
Floor surface:  GID 16
Fill below:     GID 28
Left wall:      GID 18  (placed on column 0 or leftmost boundary)
Right wall:     GID 17  (placed on last column or rightmost boundary)
Empty/air:      GID 0
```

---

## 5. `<layer>` Element (Tile Layers)

```xml
<layer id="1" name="wg-tiles" width="20" height="20">
 <properties>
  <property name="render" type="bool" value="true"/>
 </properties>
 <data encoding="csv">
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,
28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28
</data>
</layer>
```

| Attribute | Required | Description |
|-----------|----------|-------------|
| `id` | yes | Unique layer ID (positive integer). |
| `name` | yes | Layer name. `"wg-tiles"` = collision layer. |
| `width` | yes | Must match `<map>` width. |
| `height` | yes | Must match `<map>` height. |
| `opacity` | no | 0.0-1.0, defaults to 1. |
| `visible` | no | `1` or `0`, defaults to 1. |

### CSV Data Rules

- `<data encoding="csv">` -- always use CSV encoding in this project.
- Data is a flat list of GIDs, row by row, left-to-right, top-to-bottom.
- Each row has exactly `width` comma-separated values.
- Every row **except the last** ends with a trailing comma `,` then a newline.
- The **last row** has NO trailing comma, followed immediately by a newline and `</data>`.
- Total values = `width * height`.
- GID `0` = empty/no tile.
- No spaces between values.

### Project Layer Names

| Layer Name | Purpose | `render` property |
|------------|---------|-------------------|
| `wg-tiles` | Collision geometry. Parsed into `SolidTile` structs. | `true` (also renders visually) |
| *(other names)* | Visual-only decoration layers. | `true` to render, `false` or omitted to skip |

The game engine (`assets/assets.go`) iterates all layers: any layer with `render=true` gets drawn to the background image. Only the `wg-tiles` layer is parsed for collision data.

---

## 6. `<imagelayer>` Element

Used for background images behind tile layers.

```xml
<imagelayer id="6" name="bg" opacity="0.3">
 <image source="background/bg-cyberpunk-large.png" width="4064" height="2048"/>
 <properties>
  <property name="render" type="bool" value="true"/>
 </properties>
</imagelayer>
```

| Attribute | Description |
|-----------|-------------|
| `id` | Unique layer ID. |
| `name` | Layer name. |
| `opacity` | 0.0-1.0. Controls transparency. |
| `offsetx`, `offsety` | Pixel offset for positioning. |

The `<image>` child has `source` (relative path), `width`, and `height`.

---

## 7. `<objectgroup>` Element

Contains `<object>` elements. Used for spawn points, zones, paths, and metadata markers.

```xml
<objectgroup id="2" name="Connections">
 <object id="1" name="exit_right" x="272" y="224" width="48" height="48">
  <properties>
   <property name="edge" value="right"/>
   <property name="slot" type="int" value="0"/>
  </properties>
 </object>
</objectgroup>
```

### `<objectgroup>` Attributes

| Attribute | Required | Description |
|-----------|----------|-------------|
| `id` | yes | Unique layer ID (shares ID space with `<layer>`). |
| `name` | yes | Group name. Determines how the engine parses it. |
| `color` | no | Display color in Tiled editor (e.g. `"#ff2600"`). |
| `opacity` | no | 0.0-1.0, defaults to 1. |
| `visible` | no | `1` or `0`, defaults to 1. |

### `<object>` Attributes

| Attribute | Required | Description |
|-----------|----------|-------------|
| `id` | yes | Unique object ID (positive integer, unique across ALL objects in the map). |
| `name` | no | Human-readable name. Defaults to `""`. Always add one for clarity. |
| `type` | no | Object type/class string. Used for fire types (`"fire_pulsing"`, `"fire_continuous"`). |
| `x` | yes | X position in **pixels** from map left. |
| `y` | yes | Y position in **pixels** from map top. |
| `width` | no | Width in pixels. Defaults to 0 (point object). |
| `height` | no | Height in pixels. Defaults to 0 (point object). |
| `rotation` | no | Rotation in degrees clockwise. Defaults to 0. |

**Object types by shape:**
- **Point object:** No `width`/`height` (or both 0). Used for spawn positions.
- **Rectangle object:** Has `width` and `height`. Used for zones, connection markers.
- **Polyline object:** Contains a `<polyline points="..."/>` child. Used for patrol paths.
- **Point marker:** Contains a `<point/>` child. Used for fire obstacles.

### Project Object Group Names

The engine (`assets/assets.go`) and chunk loader (`procgen/chunk.go`) recognize these names:

| Group Name | Parsed By | Object Format |
|------------|-----------|---------------|
| `PlayerSpawn` | `assets.go` | Point at (x,y). Property: `spawnPoint` (string or int). |
| `EnemySpawn` | `assets.go` | Point at (x,y). Properties: `enemyType` (string), `pathName` (string). |
| `PatrolPaths` | `assets.go` | Named polyline objects. `<polyline points="dx1,dy1 dx2,dy2"/>` |
| `DeadZones` | `assets.go` | Rectangle (x, y, width, height). No custom properties needed. |
| `Checkpoint` | `assets.go` | Rectangle. Property: `checkpointID` (float). |
| `Obstacles` | `assets.go` | Point with `type="fire_pulsing"` or `"fire_continuous"`. Property: `Direction` (string). |
| `Messages` | `assets.go` | Point at (x,y). Property: `message_id` (float). |
| `FinishLine` | `assets.go` | Rectangle (x, y, width, height). No properties needed. |
| `Connections` | `procgen/chunk.go` | Rectangle. Properties: `edge` (string), `slot` (int). |
| `EnemySlots` | `procgen/chunk.go` | Rectangle at (x,y) with `width` = platform extent. |
| `HazardSlots` | `procgen/chunk.go` | Rectangle. Property: `hazard_type` (string: `"fire"` or `"deadzone"`). |

---

## 8. Procgen Chunk Conventions

Chunks are `.tmx` files in `assets/chunks/`. They use the same tileset as campaign levels but follow additional conventions for procedural assembly.

### Required Map Properties

```xml
<properties>
 <property name="chunk_id" value="combat_01"/>
 <property name="biome" value="cyberpunk"/>
 <property name="difficulty" type="int" value="2"/>
 <property name="tags" value="combat"/>
 <property name="min_enemies" type="int" value="2"/>
 <property name="max_enemies" type="int" value="4"/>
</properties>
```

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `chunk_id` | string | **yes** | Unique identifier. Convention: `{tag}_{number}`. |
| `biome` | string | no | Biome name. Defaults to `"default"`. |
| `difficulty` | int | no | 1-5 scale. Defaults to 1. |
| `tags` | string | yes | Comma-separated: `combat`, `traversal`, `break`, `start`, `exit`, `vertical`, `hazard`. |
| `min_enemies` | int | no | Minimum enemies for dynamic placement. |
| `max_enemies` | int | no | Maximum enemies for dynamic placement. |

### Tileset Reference

Chunks live in `assets/chunks/`, so the tileset path is always:

```xml
<tileset firstgid="1" source="../levels/tilesets/cyberpunk-tiles.tsx"/>
```

### Standard Chunk Sizes

| Size | Width (tiles) | Height (tiles) |
|------|---------------|----------------|
| Small | 20 | 20 |
| Medium | 40 | 20 |
| Large | 60 | 20 |

### Connection Points

Connection objects mark where chunks attach to adjacent chunks during assembly.

```xml
<objectgroup id="2" name="Connections">
 <object id="1" name="entry_left" x="0" y="224" width="48" height="48">
  <properties>
   <property name="edge" value="left"/>
   <property name="slot" type="int" value="0"/>
  </properties>
 </object>
 <object id="2" name="exit_right" x="592" y="224" width="48" height="48">
  <properties>
   <property name="edge" value="right"/>
   <property name="slot" type="int" value="0"/>
  </properties>
 </object>
</objectgroup>
```

**Rules:**
- `edge`: `"left"`, `"right"`, `"top"`, or `"bottom"`.
- `width`: Opening width in pixels (typically 48 = 3 tiles).
- `height`: Opening height in pixels (typically 48 = 3 tiles). Always set this so the rectangle is visible in Tiled.
- `slot`: Integer slot index (for multiple connections on the same edge).
- `x` position: For left connections, `x=0`. For right connections, `x = (map_width_px - width)`.
- `y` position: Represents the **top of the opening**. For a floor at row 17 (y=272) with a 48px opening, `y=224`.
- Always give connection objects a descriptive `name` (e.g. `"entry_left"`, `"exit_right"`).

### Floor and Wall Conventions

**Floor must extend to both chunk edges** where connections exist. When chunks are assembled side by side, the player walks continuously from one chunk's floor onto the next. A gap at the edge creates a pit between chunks.

```
CORRECT:  chunk A floor extends to right edge --> chunk B floor starts at left edge
          [...16,16,16,16] | [16,16,16,16...]

WRONG:    gap at chunk boundary --> player falls
          [...16,16,0,0] | [16,16,16,16...]
```

**Boundary walls** are only needed on edges WITHOUT connections:
- `start` chunk: Left wall (GID 18 on column 0), floor extends to right edge.
- `exit` chunk: Right wall (GID 17 on last column), floor extends from left edge.
- Middle chunks: No boundary walls. Floor extends full width.

### Spawn Position Y Coordinates

Spawn Y values represent the **top** of the entity. To place an entity standing ON a floor:

```
spawn_y = floor_y - entity_collision_height
```

| Entity | Collision Height | Floor at row 17 (y=272) | Spawn Y |
|--------|-----------------|------------------------|---------|
| Player | 40px | 272 | **232** |
| Guard | 40px | 272 | **232** |
| LightGuard | 36px | 272 | **236** |
| HeavyGuard | 44px | 272 | **228** |

---

## 9. Complete Chunk Template

A minimal valid chunk with left+right connections:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<map version="1.10" tiledversion="1.11.2" orientation="orthogonal" renderorder="right-down" width="20" height="20" tilewidth="16" tileheight="16" infinite="0" nextlayerid="4" nextobjectid="5">
 <properties>
  <property name="chunk_id" value="CHANGE_ME"/>
  <property name="biome" value="cyberpunk"/>
  <property name="difficulty" type="int" value="1"/>
  <property name="tags" value="CHANGE_ME"/>
  <property name="min_enemies" type="int" value="0"/>
  <property name="max_enemies" type="int" value="0"/>
 </properties>
 <tileset firstgid="1" source="../levels/tilesets/cyberpunk-tiles.tsx"/>
 <layer id="1" name="wg-tiles" width="20" height="20">
  <properties>
   <property name="render" type="bool" value="true"/>
  </properties>
  <data encoding="csv">
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,16,
28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,
28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28,28
</data>
 </layer>
 <objectgroup id="2" name="Connections">
  <object id="1" name="entry_left" x="0" y="224" width="48" height="48">
   <properties>
    <property name="edge" value="left"/>
    <property name="slot" type="int" value="0"/>
   </properties>
  </object>
  <object id="2" name="exit_right" x="272" y="224" width="48" height="48">
   <properties>
    <property name="edge" value="right"/>
    <property name="slot" type="int" value="0"/>
   </properties>
  </object>
 </objectgroup>
</map>
```

---

## 10. Validation Checklist

Before committing a `.tmx` file, verify:

- [ ] `<?xml version="1.0" encoding="UTF-8"?>` header present
- [ ] `<map>` has all required attributes (`version`, `orientation`, `width`, `height`, `tilewidth="16"`, `tileheight="16"`)
- [ ] `nextlayerid` > max layer/objectgroup `id` used
- [ ] `nextobjectid` > max object `id` used
- [ ] All object `id` values are unique across the entire file
- [ ] All layer/objectgroup `id` values are unique across the entire file
- [ ] `<tileset>` appears before any `<layer>` or `<objectgroup>`
- [ ] CSV data has exactly `width * height` values
- [ ] Every CSV row has exactly `width` values
- [ ] Trailing comma on every CSV row except the last
- [ ] No trailing comma on the last CSV row
- [ ] GID values correspond to actual tiles in the tileset (see GID table above)
- [ ] `<properties>` with `type="int"` for integers, `type="bool"` for booleans
- [ ] String properties either omit `type` or use `type="string"`
- [ ] Object `name` attributes set for readability
- [ ] Connection objects have both `width` and `height` set (visible in Tiled)

### Chunk-specific checks

- [ ] Map property `chunk_id` is set and unique
- [ ] Map property `tags` is set (comma-separated valid tag names)
- [ ] Floor (GID 16) extends to both edges where connections exist
- [ ] Boundary walls (GID 18 left, GID 17 right) only on edges without connections
- [ ] Spawn/slot Y positions account for entity height above the floor surface
- [ ] Connection `y` = floor_y - opening_height (typically `272 - 48 = 224` for standard layouts)

---

## 11. Common Mistakes

| Mistake | Symptom | Fix |
|---------|---------|-----|
| Missing `type="int"` on integer property | `GetInt()` returns 0 | Add `type="int"` |
| Trailing comma on last CSV row | Parse error or extra empty tile | Remove the comma |
| Missing comma between CSV values | Parse error | Add comma |
| Wrong GID (e.g. using tile_id instead of GID) | Wrong tile renders, or empty | GID = firstgid + tile_id = 1 + tile_id |
| Object `width`/`height` omitted | Invisible in Tiled, `Width`=0 in parsed struct | Always set for rectangle objects |
| Floor gap at chunk edge | Player falls between chunks | Extend floor to full width |
| Spawn Y at floor level instead of above it | Entity stuck in floor | spawn_y = floor_y - collision_height |
| Duplicate object IDs | Undefined behavior in Tiled and parser | Keep all object IDs unique across the file |
| `nextobjectid` too low | Tiled may reuse IDs causing conflicts | Set higher than any object ID in file |
