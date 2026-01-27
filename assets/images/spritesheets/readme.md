## Spritesheet Information

### Default Frame Sizes
Each character image is:
- width is 96 pixels
- height is 84 pixels

So every frame is 96 pixels wide and 84 pixels high.

### Asset Conversion (Horizontal Strips)

To optimize rendering and simplify animation logic, several sprite sheets have been converted from grid/vertical layouts to horizontal strips using **ImageMagick**.

#### Conversion Commands

```bash
# Flame Continuous (66x47 frames)
# Extracted 25 frames from a complex grid layout
magick flame_continously_66x47.png -crop 66x47 +repage miff:- | magick -[0-6,8-14,16-22,24-27] +append flame_continuous.png

# Purple Fire (24x92 frames)
# Converted a vertical/grid layout into a single horizontal strip
magick purple_fire_24x92.png -crop 24x92 +repage +append purple_fire.png

# Fireball (42x89 frames)
# Converted a vertical/grid layout into a single horizontal strip
magick fireball_42x89.png -crop 42x89 +repage +append fireball.png
```

This ensures compatibility with the game's animation system, which expects horizontal strips where frames are indexed from left to right.
