

# Tiled Tip Sheet

Last updated: 05 Sep 2025

This is a collection of various tips to help you better utilize
[Tiled](https://www.mapeditor.org/). It is *not* intended for first-time
users. Rather, this document collects the advice and answers I often
find myself giving on the Tiled Discord and forum, so that it is easier
to find and refer to. This is not intended as a replacement for the
[Tiled
documentation](https://doc.mapeditor.org/en/stable/manual/introduction),
but as a companion to it, and will link to it where appropriate.

In addition, I've noticed that many tutorials focus too much
step-by-step processes and that users do not develop an understanding of
how Tiled's tools work and are unable to effectively use Tiled in any
scenarios that don't closely match what they saw in those tutorials. I
hope that my focus on *how to approach* various tasks in Tiled will help
you be a better-rounded Tiled user.

> [!NOTE]
> Sometimes it is helpful to use words both in their everyday meanings
> and in their Tiled-specific meanings. I will capitalize terms when
> referring to Tiled tools and features, and use all lower-case when
> using words in their everyday meaning. For example, "Tileset" means a
> tileset as used in Tiled, such as a TSX file, while "tileset" refers
> to the broader concept or to tileset images.

## Table of Contents

- [Missing Panels](#panels)
- [Map Orientation](#orientation)
  - [Hexagonal considerations](#hexagonal)
- [Misaligned Tiles](#misaligned)
- [Tile Stamps Panel](#stamps)
  - [Variations](#stamps-variations)
  - [Random Mode](#stamps-random)
- [RPG Maker Tilesets](#rpgm)
- [Terrains](#terrains)
  - [Identifying Terrains](#terrains-identifying)
  - [Terrain Set types](#terrains-types)
  - [Labelling Terrains](#terrains-labelling)
  - [Incomplete Terrains](#terrains-incomplete)
  - [Transitional Terrains](#terrains-transitional)
  - [Can't Paint with a Terrain](#terrains-nopaint)
- [Tileset Management](#tileset-management)
  - [Extending and Rearranging Tilesets](#tileset-extend)
  - [Texture Packing and Tiled](#tileset-packing)
  - [Navigating Tilesets](#tileset-search)
- [Common scenarios](#scenarios)
  - [RPG Trees](#scenarios-trees)
- [Moving Tiled Files](#broken)
  - [Embedded Tilesets](#broken-embedded)
- [Parsing Tiled Maps](#parsing)
  - [Garbled-looking Tile Layer Data](#parsing-format)
    - [Parsing the Base64-encoded formats](#parsing-base64)
  - [Very Large Tile IDs](#parsing-large)
  - [Tile IDs slightly off](#parsing-firstgid)
  - [Loading Tilesets](#parsing-tilesets)
- [Rendering Tiled Maps](#rendering)
  - [Drawing tiles from a Tileset](#rendering-tilesets)
  - [Rendering Maps](#rendering-maps)
    - [Orthogonal Map cells](#rendering-ortho)
    - [Isometric Map cells](#rendering-isometric)
    - [Staggered Map cells (Hexagonal and
      Isometric)](#rendering-staggered)
- [Scripting](#scripting)
  - [CLI](#scripting-cli)
  - [GUI](#scripting-gui)
    - [Map Editor](#scripting-mapEditor)
    - [Tileset Editor](#scripting-tilesetEditor)
  - [Scripting Caveats](#scripting-caveats)
  - [Reading and Writing Files](#scripting-io)
  - [Sharing Your Scripts](#scripting-sharing)
- [Credits](#credits)

## Missing Panels

If you close a panel or can't find some panel you need, the first place
to look is the `View` menu. At the top is the `Views and Toolbars`
submenu, from which you can enable and disable most panels.

The `Tile Stamps` panel is closed by default. This is a very useful
panel which is barely mentioned in the official Tiled documentation. You
can read more about it [here](#stamps).

## Map Orientation

In Tiled, only *maps* have orientations - orthogonal, isometric,
hexagonal. Tilesets do not have any meaningful concept of orientation,
except as a hint to help Tiled render Terrain labels correctly.

A tile is always a rectangular image, no matter the type and style of
Tileset. This means that even if you're using an isometric or hexagonal
tileset, you can't use a tileset image where the tiles are arranged on
an isometric or hexagonal grid, because Tiled will not be able to cut
out the individual tiles from that.

The map's orientation determines where these rectangles are drawn,
creating the appearance of other orientations. This is how most game
engines render non-orthogonal tiles, too.

The fact that tiles are fundamentally rectangles regardless of map
orientation has implications for Tile Objects: they're always orthogonal
rectangles, even on non-orthogonal maps! This is particularly
disorienting on isometric maps, where normal rectangular objects conform
to the isometric grid, but Tile Objects do not. If you need to have
Objects with world-space coordinates on isometric maps, it is probably
better to avoid using Tile Objects and get by with regular Objects. You
can set the tile or other graphical information using Custom Properties
instead.

### Hexagonal considerations

For the most part, using hexagonal tiles in Tiled is just like using
orthogonal and isometric tiles. However, there are some limitations and
bugs that affect hexagonal maps:

- Terrains only support four corners/edges per tile. While you can label
  a hexagonal Tileset with terrains, it will not work correctly with the
  Terrain tools on hexagonal maps.
- Automapping is buggy on Staggered maps, and all hex maps are
  staggered. This issue also affects Staggered isometric maps.
  - You can work around this bug by creating two copies of your rules,
    positioning them so that they start with different staggers (you can
    tell if the rules look the same with the map set to a Staggered
    orientation, but different when the map is set to Orthogonal, which
    is what the Automapping system sees), and using the mod and offset
    Automapping properties to have these different rules run depending
    on whether a location in the map is in a Staggered row/column or
    not. If your stagger axis is X, set `ModX` in the Map Properties to
    2, otherwise set `ModY` to 2. After you've drawn your rules, use
    `rule_options` to make one set of rules have either `OffsetX` (if
    stagger axis is X) or `OffsetY` (if stagger axis is Y) at 1, and the
    other at 0. Which rule(s) should have which may take some
    experimenting to figure out, especially if your rules are all
    different sizes. When drawing your options rectangles, do so in
    Orthogonal mode to make sure they cover the rules you actually want
    them to cover. When viewed in a Staggered orientation, the Objects
    may appear to overlap different rules. Alternatively, you can put
    the rules in separate maps and using Map Properties for everything,
    avoiding the uncertainty of Objects. You'll still need to carefully
    position the rules, however (you can accomplish this quickly by
    making a copy of the first map, and adding 1 row or column at the
    top/left of the map using the Resize Map feature).
- Before Tiled 1.7.2, multi-tile stamps did not always work correctly on
  Staggered maps, including hex maps. If you're doing any work with hex
  maps, it's worth using this version or newer.

There are also two additional properties that affect the appearance of
hexagonal Maps:

- Stagger Axis affects whether tiles are staggered in X or Y. In more
  practical terms, this determines whether your hexagons are
  pointy-topped (Y) or flat-topped (X). Choose the one appropriate to
  your tiles' artwork.
- Hex side length is what makes hexagons hexagonal, it determines the
  length of the horizontal or vertical sides of the hexagon. At 0, your
  hexagons will look like isometric tiles. Choose a value that matches
  your tiles' horizontal or vertical sides, though even if you count the
  pixels, you may need to do a little trial and error to get the tiles
  lining up perfectly, as there's often intended to be some minor
  overlap between hexagonal tiles.

## Misaligned Tiles

Many isometric and hexagonal tilesets feature tiles that have some depth
to them. If you set your map's tile size to be the same as the tileset's
tile size with such tiles, they will not align correctly. This is
because the actual functional surface of these tiles is smaller than the
space the tiles take up in the image.

![Isometric map with grass-topped ground tiles. The tiles are misaligned, and the sides are visible where they should not be. - Incorrectly aligned isometric tiles due to an incorrectly
set map tile size.](https://eishiya.com/articles/tiled/images/isometric1.png)

*Incorrectly aligned isometric tiles due to an incorrectly
set map tile size.*

![The same map and ground tiles, but with the tops aligning neatly, forming a continuous ground surface. - The same tiles with the map tile size set
correctly.](https://eishiya.com/articles/tiled/images/isometric2.png)

*The same tiles with the map tile size set
correctly.*

To fix this, the map's Tile Height needs to be set to a smaller value in
`Map > Map Properties`. The tile height in the *Map* should be the
height of the *flat surface* of the tile, not including its depth, since
the depth should not contribute to tiling a continuous flat area with
the tiles. You should use the most basic flat tile in your tileset to
determine the Tile Height, do not use tiles that slope or have details
that obscure the tile's top surface.

For isometric maps, you can measure the height of the surface of the
tile in an image editor, or reduce the Tile Height of the Map in Tiled
until the tiling looks correct. For pixel art tiles that don't have
anything sticking out their sides, the correct height is almost always
1/2 the width, so that's a good value to start with.

For hexagonal maps, you will also need to set the Tile Side Length
correctly to get proper tiling. The most reliable way is to measure in
an image editor. If you want to just adjust the sizes in Tiled instead
of measuring, start by reducing the Tile Height until the bottom of the
tile grid aligns with the bottom edge of the surface of the tile. After
that, adjust the Tile Side Length until the tile grid matches the
surface of the tile and the tiles tile correctly.

> [!NOTE]
> You can use the Up and Down arrow keys on your keyboard to quickly
> change the value of the properties instead of typing values in. Click
> the value to edit it, and then you can use the arrow keys. You can
> also use the small arrows that show up on the field using your mouse.

If your tiles' surface doesn't span the entire width of the tile, you'll
also need to adjust the map's Tile Width. This can happen with tilesets
that include details that go off the tile.

> [!NOTE]
> If your tileset has tiles placed haphazardly, that is, placed such
> that the top surfaces aren't always in the same place within the tile,
> you will not be able to get them to align. The only remedy is to
> adjust the tileset image(s) in an image editor so that the tiles are
> aligned consistently.

## Tile Stamps Panel

A hidden gem in Tiled is the Tile Stamps panel. It lets you save
arrangements of tiles for later reuse, and allows you to paint with
random Stamps. This panel is hidden by default, you'll need to enable it
in `View > Views and Toolbars > Tile Stamps`.

> [!NOTE]
> In Tiled, "(Tile) Stamp" doesn't *only* refer to Stamps saved in the
> Tile Stamps panel. Any time you use the Stamp Brush, which is the
> basic drawing tool in Tiled, your brush is a Tile Stamp. Any time you
> pick one or more tiles from a Tileset to draw with, that's a Tile
> Stamp. The only things different about the Stamps in the Tile Stamps
> panel is that they're saved for later use, and that random variations
> can be selected.

### Variations

A very useful feature of this panel is the ability to create Variations
of Tile Stamps. For example, you can make a tree Stamp and save a bunch
of other trees as Variations of that Stamp. Then, when you use that tree
Stamp, Tiled will pick a random tree to place. To use this feature, you
first need to create a base Stamp. Select the tile or tile arrangement
you want to make a stamp out of, and click the `Add New Stamp` button.
Then, one by one, select your variants and click the `Add Variation`
button.

> [!NOTE]
> Tiled currently has no way to change whether a given Stamp is a base
> Stamp or a Variation, so be careful when creating stamps. If you
> accidentally make a base Stamp when you want a Variation or vice
> versa, you'll need to delete and re-add it.

After you've created your Stamp with Variations, any time you paint with
the base Stamp, a random Stamp will be selected from among the base
Stamp and its Variations. If you click the arrow to the left of the
Stamp in the Tile Stamps panel, you can expand or collapse the Tile
Stamp. When expanded, you can view and edit the Variations. One of the
properties you can edit is the Probability, which lets you control how
often a particular stamp shows up relative to the others. Probability is
only meaningful for Variations. The probability listed for the base
Stamp is just the sum of the probabilities of the variations and isn't
used for anything.

If you select a Variation to draw with instead of the base Stamp, only
that Variation will be used. This means you can also use Variations as a
way to group Stamps together, rather than just for randomness.

When using Variations, you'll usually want the Stamp Brush's random mode
to be *off*, so that the entire Stamp is drawn, rather than a random
tile from it.

### Random Mode

The Tile Stamps panel can be used to improve the usefulness of the Stamp
Brush's Random mode. In Random mode, a single random tile from your
current brush (Stamp) is drawn. You can save commonly-needed tile
selections as Stamps, saving you the trouble of selecting them from your
tileset or map every time. One thing you can do with stamps that you
can't do by selecting tiles directly from your tileset is add flipped
tiles. Using a stamp that contains various flips and rotations of your
tiles allows you to randomize between those flips and rotations when you
draw using Random Mode.

![Two columns of two ladder tiles each. The right column is a horizontal mirroring of the left. - This Tile Stamp contains two ladder tiles and their
horizontal flips. Using this Stamp with Random Mode will draw a random
and randomly-flipped ladder tile.](https://eishiya.com/articles/tiled/images/stamps_random1.png)

*This Tile Stamp contains two ladder tiles and their
horizontal flips. Using this Stamp with Random Mode will draw a random
and randomly-flipped ladder tile.*

Of course, when you just need a one-off bunch of random tiles like this,
you can draw them on a map and copy+paste them from there to use with
Random Mode. You only need to use the Tile Stamps panel when you want to
save that bunch of tiles to use later.

## RPG Maker Tilesets

Some of the stock tilesets you'll find online, both free and paid, will
be designed for RPG Maker autotiling. Such tilesets are stored in a
compressed format, designed to be broken up into smaller sub-tiles and
reassembled into full-size tiles at runtime. Tiled cannot use such
tilesets directly effectively.

![A 2x3 arrangement of grass and dirt tiles in RPG Maker autotile format. - Tilesets with tiles arranged in patterns like this are
usually designed for RPG Maker autotiling and cannot easily be used
directly in Tiled.](https://eishiya.com/articles/tiled/images/rpgm.png)

*Tilesets with tiles arranged in patterns like this are
usually designed for RPG Maker autotiling and cannot easily be used
directly in Tiled.*

In theory, you could use such tilesets in Tiled by loading them in with
half the tile size they're intended for, but this will make it
troublesome to select the exact tiles you need, and they won't play very
nicely with Terrains. Instead, you should expand the tilesets by
building full tilesets out of their subtiles. These are some tools
available to automate this:

- [devium's RPG Maker to blob converter (Python script,
  multi-OS)](https://github.com/devium/tiled-autotile)
- [fmoo's RPG Maker autotile packing/unpacking tools (Windows
  executable)](https://fmoo.itch.io/autotile-packing-tools)
- [eishiya's Expand RPG Maker Autotile Tileset (Tiled
  script)](https://github.com/eishiya/tiled-expand-autotile)

I recommend the last one (full disclosure: I wrote it), as it runs
directly in Tiled (1.10.2+) and behaves similarly to the regular New
Tileset dialog. It also lets you choose whether to save the expanded
intermediate as an image (similar to the other scripts), or as a
TileMap. If you do the latter, you'll have a metatileset that still uses
the original, much smaller, source image. The disadvantage of using a
metatileset is that not many parsers/engines support them.

## Terrains

[Terrains](https://doc.mapeditor.org/en/stable/manual/terrain/) are
simple to use once you set them up, but if you don't understand *how*
the feature works, you may have difficulty correctly labelling your
tiles if the tileset is complex or arranged in an unusual way. There's
really just one thing to understand about Terrains, everything else
flows from it: **the labels tell Tiled how the tiles are intended to
connect rather than what's on them, if the labels on the adjacent sides
of two tiles match, they're allowed to placed next to each other**,
otherwise they are not. This is different from how some other level
editors approach this task, but has the benefit of being very flexible.

![Two water-and-grass tiles next to each other, shown without labels above and with labels below. The edges where they meet are entirely water. The matching edges are circled in pink. - The adjacent labels on these tiles match, so the Terrain
tools will place them together.](https://eishiya.com/articles/tiled/images/terrains_labels4.png)

*The adjacent labels on these tiles match, so the Terrain
tools will place them together.*

![Similar to the previous figure, but the right tile is different, with both grass and water along its edge, no longer a sensible match for the left tile. The mismatched edges are circled in pink. - The adjacent labels on these tiles do not match, so
the Terrain tools will never place them together like this.](https://eishiya.com/articles/tiled/images/terrains_labels5.png)

*The adjacent labels on these tiles do not match, so
the Terrain tools will never place them together like this.*

This concept of matching labels is applied in all four directions, the
Terrain tools will try to find an arrangement of tiles such that all the
labels on adjacent sides match perfectly. It is from these small
relationships that larger coherent terrain shapes are built up.

Before you can label your tileset with Terrains, you need to understand
how the tiles should be used, how they should connect to each other. If
you've just downloaded a new tileset, take the time to play around with
it by manually placing tiles before you start labelling it with
Terrains. Understanding how your tileset works will make labelling it
much easier. Consider not only which arrangements of tiles you're likely
to use, but also how you may split your map across layers. For example,
many RPG-style tilesets have plateaus where the top and sides of the
plateau has transparency, which means plateaus should be drawn on a
layer above the base terrain, and those Terrains should therefore
transition to empty rather than to the ground Terrains.

> [!NOTE]
> In the interest of performance, Tiled does not do an exhaustive search
> for a valid arrangement of tiles. So, Tiled will sometimes make
> mistakes and leave you with an invalid arrangement if you're using
> incomplete Terrains. Because of this, it's best to make your Terrains
> as simple and complete as possible. It's often better to handle the
> complex transitions with Automapping that runs after you paint with
> Terrains than to try to do *everything* with Terrains.

> [!NOTE]
> Not every tileset is well-suited to being used with Terrains. Some
> tilesets have complex, multi-tile transitions that can't be easily
> represented with Terrains, some rely on sharp boundaries and only have
> a small number of transitions, some rely on each tile being used in
> several contexts. Such tilesets are better used with
> [Automapping](https://doc.mapeditor.org/en/stable/manual/automapping/)
> than Terrains.
>
> Some tilesets can be used with Terrains, but have only a small subset
> of transitions that they support, which will lead to broken results if
> you try to draw unsupported situations with them. It's important to
> know your tileset's limitations.
>
> And lastly, some tilesets are so irregular that neither Automapping
> nor Terrains are a good fit, their tiles are best placed manually.

### Identifying Terrains

To avoid accidentally mislabelling some tiles, look through your tileset
and figure out which tiles are part of which terrains, *before* you
start labelling. Sometimes things that look different may actually be
part of the same terrain. For example, a tileset may look like it has
water, sand, and grass terrains, but the sand may just be decoration at
all water-grass edges, so you wouldn't actually need a separate sand
terrain, just water and grass. As a starting point, look for those tiles
are filled with some terrain (e.g. all-water, all-grass tiles), and see
if there are tiles that could serve as transitions between those tiles
and other filled tiles.

Some Tilesets have multiple independent sets of terrain in them, that
is, groups of terrains that connect with other terrains in their group,
but not with terrains in other groups. For example, a tileset may have
some interior wall and floor tiles that can be used as Terrains, but
which have no way to connect to the exterior Terrains. I recommend
having these as *separate Terrain Sets*, so that you don't have to look
through the entire list of Terrains every time you want to find the one
you want.

### Terrain Set types

When you [create a new Terrain
Set](https://doc.mapeditor.org/en/stable/manual/terrain/#creating-the-terrain-set),
you will need to choose a Terrain Set type. If you don't know what type
of terrain you have, try to build the smallest shapes of each terrain
that you can with the tiles, not counting 1x1 single-tile "islands". If
it's a filled 2x2 shape connected at the corners, it's a Corner-based
terrain. If it's a 2x1 or 1x2 line, it's an Edge-based terrain. If you
can make both, it's a Mixed terrain.

![Four-tile shape using a corner terrain set. - Filled four-tile shapes like this are only possible with
Corner (and Mixed) Terrains.](https://eishiya.com/articles/tiled/images/terrains_cornerIsland.png)

*Filled four-tile shapes like this are only possible with
Corner (and Mixed) Terrains.*

![Horizontal and vertical two-tile shapes using an edge terrain set. - Two-tile shapes like this are only possible with Edge (and
Mixed) Terrains.](https://eishiya.com/articles/tiled/images/terrains_edgeIslands.png)

*Two-tile shapes like this are only possible with Edge (and
Mixed) Terrains.*

Because most tilesets are [incomplete](#terrains-incomplete), you may
find yourself unable to make these shapes in every combination of
terrains, but only some of them. In particular, with Mixed tilesets,
it's very common that you can make the two-tile shapes with one terrain,
but not with the terrain around it. The island tileset above, for
example, can make two-tile islands but not two-tile lakes. If you
suspect you may be dealing with an incomplete Mixed tileset (i.e. you
can make four-tile corner-linked shapes, but have way more tiles than
would be needed for a corner-based terrain), try making the two-tile
shapes in both configurations - terrain A inside and terrain B outside,
and vice versa. If at least one of them works, you've definitely got a
Mixed Terrain Set.

> [!NOTE]
> It's common for a single tileset to contain multiple sets of terrains
> with different types. For example, top-down tilesets usually have
> Corner-based ground terrain, but may include Edge-based details like
> roads and fences. For this reason, you should check the type for all
> the different terrains you want to use.
>
> The types and labels for each Terrain Set are independent, so you can
> set up the labels in each Terrain Set in whatever way is most useful
> for that set, the label for a particular tile can be different in each
> Terrain Set.

### Labelling Terrains

Once you've created your Terrain Set, you can start labelling. Place
each terrain colour at each corner and/or edge (depending on your
terrain type) where that terrain occurs. The [Terrain
documentation](https://doc.mapeditor.org/en/stable/manual/terrain/) has
more information on the UI for labelling and using Terrains.

> [!NOTE]
> If you're working with an isometric tileset, you can set the
> Orientation in the Tileset Properties to "Isometric" and set the grid
> size to match your tiles' visuals, so that the Terrain labels more
> closely match the art.

My usual advice for labelling Terrains is to put each label colour on
those corners/edges of each tile where that terrain or material is
present. For many tilesets, especially top-down ones, the labels will
match the art well.

![Side-by-side comparison of several tiles and their Terrain labels. The blue water labels neatly overlay the water parts of the tiles, while the green ground labels neatly overlay the grassy parts of the tiles. - Part of a Mixed Terrain Set, with the art and the labels
side by side.](https://eishiya.com/articles/tiled/images/terrains_labels1.png)

*Part of a Mixed Terrain Set, with the art and the labels
side by side.*

> [!NOTE]
> Choose your Terrains' colours so that they resemble or evoke the
> terrain they represent, such as blue for water and green for grass,
> like in the example above. This can make labelling less confusing. You
> can customize the colour for your Terrains by right-clicking the
> Terrain name in the list of Terrains and choosing "Pick Custom Color".
>
> Avoid using colours that exactly match the tile art, choose more
> saturated colours instead, so that they stand out from the art.

However, it can be harder to see where each terrain goes if the art on
the tiles has different proportions from the Terrain labels in Tiled.

![Part of a sidescroller tileset where the entire tile is covered by the ground, without Terrain labels. - These sidescroller ground terrain tiles should transition to
empty space, but the entirety of each tile is covered in
ground.](https://eishiya.com/articles/tiled/images/terrains_labels2.png)

*These sidescroller ground terrain tiles should transition to
empty space, but the entirety of each tile is covered in
ground.*

![The same sidescroller tiles, overlaid with red Terrain labels. The labels only cover part of the tile. - The ground in the tile art reaches the tile edges, but the
labels do not.](https://eishiya.com/articles/tiled/images/terrains_labels3.png)

*The ground in the tile art reaches the tile edges, but the
labels do not.*

**The key thing to keep in mind is that when you label Terrains, what
you're doing is telling Tiled *how the tiles should connect* to one
another.** It doesn't matter what the labels look like, what's important
is that the edges/corners of tiles that can be placed together have
matching labels. In the above example, the outermost parts of the tiles
that transition to empty space are left empty (unlabelled), and the rest
is filled with the ground Terrain.

When labelling Terrains, take care to stay focused on the individual
tiles and how they may connect with other tiles. Do not fall into the
trap of looking at the overall shapes your labels make across
neighbouring tiles. While there are some common arrangements of tiles in
tilesets, such as the 3x3 block of corner-based terrains, tiles can be
arranged in any way and you can't rely on their arrangement. Pay
attention to the tiles instead.

> [!NOTE]
> Some tilesets include 1x1 "island" tiles of some of their terrains.
> You can see two such tiles in the example above, the two unlabelled
> tiles. The top right one is ground on all sides and empty in the
> middle, the bottom left one is empty on all sides and ground in the
> middle. Since Tiled has no way to mark the middles of tiles with
> Terrains, these tiles are functionally identical to the all-ground and
> empty tile, respectively. Such tiles should be placed by hand.
>
> However, it is a good idea to label them where possible, so that if
> you do place them manually and then edit the surrounding tiles with
> the Terrain tools, the tools will know how to connect other tiles to
> them. In order to prevent those tiles from showing up automatically as
> random variants of the regular tiles with those same labels, set their
> Probability to 0 in the Tile Properties.

> [!NOTE]
> When labelling a Mixed Terrain Set, it's common that some of the
> Terrains are only really important at Corners, while others are only
> important at Edges. In these cases, only label those! It's perfectly
> fine to leave some Corners or Edges unlabelled if it doesn't help you
> specify how the tile should be used. Label only what's important,
> don't label anything that doesn't *need* a label.
>
> When Tiled detects that a given Terrain in a Mixed Terrain Set is only
> used at the Corners or the Edges, it'll snap to those parts of the
> tiles when you paint with that Terrain, making the Mixed Terrain Set
> easier to use.

### Incomplete Terrains

Most tilesets out there are incomplete. Even those featuring just one
set of corner-based terrains, something that normally only requires 15
or 16 tiles, are usually missing tiles. Tiled can handle incomplete
Terrains to some degree, and most common types of incomplete tilesets
will work well. However, in some cases, Tiled will have no idea what to
do, and will either not let you draw at all, or will draw an explosion
of garbage. If you find yourself facing these two results too
frequently, consider expanding your tileset, those missing tiles may
actually be important for your maps.

For Corner Terrain Sets, it's very common to leave out the two tiles
where opposite corners match. This means that you can't draw very tight,
snaking shapes with those tiles, and should instead draw your shapes
spaced out. If you run into issues using Corner Terrains, check if
you're missing these two tiles, adding them can help.

![Two tiles. The left tile is mostly grass and has water in the top right and bottom left. The right tile is similar, but has water in the top left and bottom right. - Tiles similar to these, with matching opposite corners, are
often missing from Corner-based tilesets.](https://eishiya.com/articles/tiled/images/terrains_incomplete1.png)

*Tiles similar to these, with matching opposite corners, are
often missing from Corner-based tilesets.*

Edge Terrain Sets for things like fences often only include those tiles
where only one or two of the edges are the "fence" and the rest are the
surrounding terrain. This means that the "fences" can't branch, as that
would require tiles where three of the edges are "fence". If only tiles
where two of the four edges are "fence" are included, the fence can only
make closed loops, it has no way to terminate to make an opening.

A complete Mixed Terrain Set with two Terrains would require 256 tiles.
Unsurprisingly, it's very rare to find such a tileset, as that's a lot
of art to produce, and most of those tiles aren't that useful. Instead,
it's far more common to use the [47-tile "Blob"
subset](https://web.archive.org/web/20230101/cr31.co.uk/stagecast/wang/blob.html).
This style of tileset allows for a large variety of terrain with only a
fraction of the tiles.

One very common mistake involving deliberately incomplete tilesets is to
try to draw on an empty layer with a Terrain that has no way to
transition to empty tiles. Tiled will not let you draw anything at all
since there's nothing it can draw, making the Terrain Brush appear to
not work. The solution is to fill your layer with the base tile for one
of the Terrains before you start drawing with the Terrain Brush, so that
there's something valid for your desired Terrain to transition to.

Tilesets not designed with Terrains/automapping in mind often require
incomplete Terrains to properly label, and the incomplete terrains often
give Tiled trouble, sometimes to the point of barely being useable.
Automapping can be a better option in these cases. If you are an artist
designing tilesets, I highly recommend getting to know the Terrain tools
in Tiled and making your tilesets work with them. Basic two-colour
Terrains in Tiled will also work well with Unity and Godot's autotiling
tools.

### Transitional Terrains

Not every tileset fits nicely with Tiled's concept of Terrains. You
can't always just add labels that tell Tiled, "here's Terrain A, here's
Terrain B", but that doesn't mean you can't make those tiles work with
Terrains. Sometimes you just have to get creative.

A not uncommon scenario is when two sets of tiles have corresponding
edges, but no actual transition tiles.

![Small room made out of tiles in a dark void. The bricks are lighter around the perimeter of the room. - These sidescroller tiles do not include any transition tiles
between the wooden back wall and the brick side walls, but the two
should clearly be connected: the side walls have light edges, and the
back wall has shading where it meets the side walls.](https://eishiya.com/articles/tiled/images/terrains_transitional1.png)

*These sidescroller tiles do not include any transition tiles
between the wooden back wall and the brick side walls, but the two
should clearly be connected: the side walls have light edges, and the
back wall has shading where it meets the side walls.*

If you label the edges of one of these as being the other terrain, then
it won't connect to the correct Tiles of the other terrain. Instead, you
need to tell Tiled that the edges of one link up with the edges of
another. And the mechanism Terrains have to do that is... another
Terrain.

![3x3 sets of tiles of bricks and wooden back wall, without labels. - Some brick side wall and back wall tiles from the tileset
used for the room above.](https://eishiya.com/articles/tiled/images/terrains_transitional2.png)

*Some brick side wall and back wall tiles from the tileset
used for the room above.*

![The same tiles, overlaid with Terrain labels: purple for the inside part of the brick walls, blue for the back wall, and orange for the light outer part of the brick walls and shaded parts of the back wall. - Labels for those tiles. The orange Terrain is the
transitional Terrain that tells Tiled that the shadows on the back wall
should connect to the light edges of the brick side walls.](https://eishiya.com/articles/tiled/images/terrains_transitional3.png)

*Labels for those tiles. The orange Terrain is the
transitional Terrain that tells Tiled that the shadows on the back wall
should connect to the light edges of the brick side walls.*

### Can't Paint with a Terrain

If you make a terrain but Tiled won't place any tiles when you try to
use the Terrain Brush, the most likely culprit is that your terrain has
no way to transition to whatever is already on the layer. This commonly
occurs when painting on an empty layer with a terrain that has no
transitions to empty. Try filling the layer with a solid tile from the
terrain set you want to use, and then painting with that.

## Tileset Management

Although Tiled generally makes working with any number of tilesets easy,
a few concerns come up repeatedly.

### Extending and Rearranging Tilesets

If you're building your tileset as you go, you'll likely need to add
tiles to your tileset(s) after initially creating them. With an Image
Collection, this is trivial, you can use the Add Tiles button in the
Tileset Editor to add your new tile images to the tileset. When adding
tiles to a "Based on Tileset Image" Tileset, things can go wrong.

"Based on Tileset Image" Tilesets assign tile IDs to each tile based on
its sequential location in the image, based on the tile size and spacing
you set. The tiles are numbered in rows from top to bottom, left to
right. This means that if the width of the tileset image changes and
more tiles can fit in each row, the IDs of the tiles under the first row
will all change. Since Tiled Maps reference tiles by their IDs, this
means that existing maps using that tileset will refer the wrong tiles
after the tiles in the Tileset are renumbered.

> [!NOTE]
> If you resize your tileset image while its Tileset is open in Tiled,
> Tiled will offer to adjust the tile IDs in open maps to keep the maps
> looking the same as before. This is often enough to make simple width
> changes work, and the tips below may be unnecessary for you.
>
> However, you only get this offer once per size change, so make sure
> you have *all* the relevant maps open in Tiled before you change your
> Tileset's width. If you end up adjusting some maps but not others, it
> gets much more difficult to fix them. Because it's easy to forget to
> open all maps and because for large projects, it may be impossible to
> have all of them open at once, I personally avoid relying on this
> Tiled feature. But if you can keep all your Tiled files open at all
> times, this Tiled feature is the easiest way to deal with changing
> Tileset widths.

While map corruption from changed tile IDs *can* be remedied by
replacing the incorrect tiles with the correct ones and there's even [a
script to
help](https://github.com/eishiya/tiled-scripts/blob/main/MassReplaceTiles.js),
it requires you to know which tile corresponds to which correct tile,
which is tedious to figure out. It is best to avoid renumbering existing
tiles in the first place. There are ways avoid issues when adding tiles
to a tileset:

- Extend your tileset image only downwards, never changing its width,
  and never rearrange existing tiles in the tileset. This will add IDs
  for the newly added tiles without changing the old ones. This approach
  is best if you size your tileset image to the maximum width you expect
  to need from the beginning. If the tileset is narrow, then it would
  have to remain narrow even as it grows very tall, which may be
  inconvenient.

- Use multiple tilesets. Instead of extending an existing tileset,
  consider putting the new tiles in a separate tileset. This can help
  with navigating large collections of tiles as well. However, it does
  have the disadvantage that Terrains only recognise tiles within their
  tileset, there's no way to have meaningful Terrain connections between
  tilesets. Using multiple tilesets can also lead to performance issues
  due to texture swapping, unless texture packing is used.

If you need to *rearrange* a tileset or insert new tiles between old
ones and don't have empty space you can use, the above approaches will
not be sufficient. To rearrange tilesets, you only really have two
options:

- Add the tiles at the end, and use Tiled's Rearrange Tiles tool to
  display them in a more convenient location in the tileset. This can
  throw off the arrangement of other tiles however, so it may take a lot
  of work to arrange your tileset to be just as you want it.

- Rearrange your tileset image, and then adjust your Maps to use the new
  tile locations. This is easiest to do with the [Mass Replace
  Tiles](https://github.com/eishiya/tiled-scripts/blob/main/MassReplaceTiles.js)
  script, especially if you save your rearranged image as a new image
  and a new Tileset. Having both the old and new tilesets available
  should make it much easier to create the mapping from old to new tiles
  that the script needs.

### Texture Packing and Tiled

Rendering usually performs better when swapping textures in and out of
memory is reduced. For this reason, it's common to want to pack one's
tiles into a texture atlas, possibly alongside other (non-tile) images.
It is often tempting to use the packed tileset as the Tileset image in
Tiled, but you should consider the downsides before you do that:

- If you add or remove tiles, your packed tileset may end up with its
  tile IDs no longer representing the same tile, breaking your maps (see
  [Extending and Rearranging Tilesets](#tileset-extend) above). This is
  especially likely if you use an automatic texture packer instead of
  arranging the atlas by hand.

- If you're using an automatic texture packer, you'll have to run it
  every time you want to add/remove tiles.

- A large tileset containing all your tiles may be inconvenient to
  navigate and find tiles in, compared to multiple smaller tilesets.

- "Based on Tileset Image" Tilesets require the tiles to be in a grid
  where every tile is the same size, so you can't trim the empty space
  or position the tiles as tightly as possible, so a Tiled-compatible
  texture atlas will often take up more memory than it would otherwise.

If your engine/framework supports texture atlases, consider making use
of that feature! Instead of using the packed texture in Tiled, use your
source images in Tiled, and then load the pixel data from the atlas
instead of the source images at runtime. This approach allows you to
keep unrelated tiles in separate Tilesets, while taking advantage of all
the benefits of texture atlases at runtime, including tight packings
that would make the atlas unusable in Tiled. If your source images were
single tiles, you can use those in Tiled by placing them in a Collection
of Images Tileset.

If you can't use a texture atlas "properly", you can at least avoid
having the tiles being positioned differently at different times by
using a texture packer like
[Atlased](https://witnessmonolith.itch.io/atlased), which allows you to
add additional images to an atlas without moving the existing ones, or
by arranging your tilesets manually in an image editor.

In the future, Tiled will likely support a new Tileset type where tiles
IDs do not change when the tileset is resized (see [issue
\#2863](https://github.com/mapeditor/tiled/issues/2863)). At that point,
using automatically-packed tilesets should become more convenient, and
it would allow using tilesets where the tiles don't follow a consistent
grid. Texture packers changing individual tiles' positions may still be
an issue, but as long as the atlas metadata contains some consistent
identifier (such as a source file name), it should be possible to use
even rearranged tilesheets without too much trouble. You can technically
already achieve all this in Tiled using Collections of Images and
setting each tile's image rect, but there's no convenient UI for it, and
you'd probably need to write some Tiled scripts to help.

### Navigating Tilesets

Many games require thousands of tiles. Whether you pack them into one
tileset or spread them across many, you may find yourself struggling to
quickly find the exact tile or tileset you need. Tiled has a few
features to help.

The tabbed list in the Tilesets view can be slow to navigate if you have
many tilesets. Don't miss the down arrow that opens a menu that lists
all your tilesets. If you're looking to open a tileset for editing, the
[Open File in Project
action](https://doc.mapeditor.org/en/stable/manual/projects/#opening-a-file-in-the-project)
allows you to search for a file to open by name.

If you have all your tiles in one big sheet, or have so many tilesets
that even the vertical menu is unhelpful, consider [saving
stamps](#stamps) with commonly-needed tiles from different regions of
the tileset. This way you can access those tiles without digging through
your tilesets, and the Tile Stamps panel shows a visual preview of them.
Selecting a Tile Stamp will navigate to the appropriate tileset and
scroll to the selected tile(s). The Tile Stamps panel can be searched by
stamp name, so if you name your stamps appropriately, you can find
everything quickly.

If you prefer a more visual sort of shortcut to your tiles/tilesets, you
can make a scratch pad map, a Tilemap with commonly used arrangements of
tiles on it. Using it will require opening it, you can't use it while
editing another map (unless you write a script to display it; no such
script exists to my knowledge). In the future, Tiled may get a feature
to display such scratch pad maps as a side panel, similar to the
Tilesets view ([issue
\#3288](https://github.com/mapeditor/tiled/issues/3288)).

## Common Scenarios

Certain scenarios or ones comparable to them come up very frequently
across many projects. This section includes some tips for dealing with
them efficiently in Tiled.

### RPG Trees

When placing trees and other tall entities in a 3/4-like view, it's
common that the player and NPCs should be able to overlap them when
walking in front of them, while being overlapped when walking behind
them. If trees are placed on a layer as a bunch of small tiles, it can
be difficult to communicate to the engine that some of those tiles
should be in front of the player, while some should be behind them.

It's also common that trees and such should be able to overlap other
trees that are behind them. However, if your trees are made out of plain
tiles, trying to place them in a single layer doesn't work correctly, as
one tree's tiles will overwrite the tiles of the trees that already
overlap those tile cells.

![Two trees, one in front of the other. The back tree is surrounded by buckets. - Trees may overlap one another, and a player or other entity
(a bucket in this case) should either overlap or be overlapped by the
tree, depending on where it's positioned relative to the
tree.](https://eishiya.com/articles/tiled/images/scenarios_trees1.png)

*Trees may overlap one another, and a player or other entity
(a bucket in this case) should either overlap or be overlapped by the
tree, depending on where it's positioned relative to the
tree.*

![Two trees, but stamping one tree has overwritten tiles of the other. - Attempting to place two trees made of multiple tiles too
close to each other on a single layer will cause the new tree's tiles to
overwrite the old tree's.](https://eishiya.com/articles/tiled/images/scenarios_trees2.png)

*Attempting to place two trees made of multiple tiles too
close to each other on a single layer will cause the new tree's tiles to
overwrite the old tree's.*

Both of these problems share the same solutions. There are two common
approaches, each with its own pros and cons:

#### Foreground and Background Layers

You can sort your tree tiles into two layers: a foreground layer that
contains the tiles that are always on top of the player, and a
background layer that contains the tiles that are always behind the
player.

![The two trees from the examples before, but the leafy crowns are highlighted. - The tree crowns are on a foreground layer (highlighted),
while the tree trunks are on a background layer below. This approach is
very common in 8- and 16-bit console games because it is well-suited to
those systems' limitations.](https://eishiya.com/articles/tiled/images/scenarios_trees3.png)

*The tree crowns are on a foreground layer (highlighted),
while the tree trunks are on a background layer below. This approach is
very common in 8- and 16-bit console games because it is well-suited to
those systems' limitations.*

Pros:

- No special handling required in the engine: simply draw the player and
  other dynamic entities between the foreground and background layers.
- Good runtime performance because there is no dynamic z-sorting, and
  all tile layer optimizations apply. Many kinds of entities can be
  included on the same background and foreground layers, so your overall
  layer count can stay low, too.
- Widely supported by engines, since it uses regular tile layers where
  each tile fits in its cell.

Cons:

- You have to manage your trees across two or more layers.
  - This can be partially mitigated by saving stamps of your trees where
    the tiles are already on the correct layers. When you use such
    stamps, Tiled will automatically put the tiles on the correct
    layers.
  - You can also set up Automapping rules to add the tree crowns on the
    foreground layer when you draw trunks on the background layer, and
    erase tree crowns from the foreground when the corresponding trunks
    are removed.
- Any part of the tree is either always behind or always in front of the
  player. If a dynamic sprite is tall enough to reach the tree's crown
  when standing in front of it, they will still be overlapped by it.
  This limitation makes this approach best-suited for games with smaller
  sprites.
- There is still a risk of "collision" and overwriting tiles if you want
  multiple trees to be very close to each other, tree foreground
  portions can only overlap only the background portion of other trees.
  You can use additional background and foreground layers to allow
  backgrounds to overlap other backgrounds and foregrounds to overlap
  other foregrounds, but that will further complicate layer management,
  and using too many layers can impact rendering performance.

#### One Tile per Tree

You create a tileset where each tree is a single tile, so it's
impossible for a part of the tree to be overwritten by something else,
and so that the tree can be z-sorted as a single entity. In Tiled, you
can use such oversized tiles on a Tile Layer, in which case they'll
stick out of their cells, or you can place them as Tile Objects on an
Object Layer. In either case, in-engine, you'll need to z-sort the trees
at runtime to know what order to render them and the other entities in.

![A group of trees tightly clustered together. One tree is still just a faded-out preview, ready to be placed. - Each tree is just one tile instead of many, so they can
overlap freely and can be placed very closely together.](https://eishiya.com/articles/tiled/images/scenarios_trees4.png)

*Each tree is just one tile instead of many, so they can
overlap freely and can be placed very closely together.*

Pros:

- Easy to edit and manage, since each tree is a single tile. If you use
  Tile Layers, you can even easily randomize which tree you place.
- If you use an Object Layer with Tile Objects without "Snap to Grid",
  you can position the trees freely, they don't have to follow the tile
  grid if you don't want them to.
- Dynamic z-sorting allows the tree to be a background or a foreground
  based on where exactly the player is relative to it.

Cons:

- You may need to put your trees and other similar props in a separate
  tileset so that each tree can be a single tile. In some cases you may
  be able to use the same tileset image as your main tileset as the
  source, just with a different tile size, but in most cases, you will
  need to make an entirely separate image or collection of images to
  serve as the source(s) for this tileset.
  - If/when Tiled gets support for Tile Atlases (see [issue
    \#2863](https://github.com/mapeditor/tiled/issues/2863)) so that you
    can define a tile out of any rectangle in your image, this issue
    will be mitigated somewhat. However, tightly packed tilesets may not
    leave the space around the narrower parts of the trees empty, and
    for those, a separate tileset would still be required.
- Z-sorting has a performance cost, and the more entities you sort, the
  more noticeable it becomes.
  - Trees are likely to be static rather than dynamic objects, which
    means they never change z-order relative to each other. Some engines
    have optimizations for sorting static entities, so their impact may
    be small.
- If you use a Tile Layer, you'll need some way to tell the engine to
  include its tiles in z-sorting rather than treating them as a single
  flat layer. Objects are usually assumed to be z-sortable dynamic
  entities, but then you might still need some way to tell the engine
  that they're z-sortable *static* entities if you want to optimize the
  z-sorting.
- When each tree is a single tile, it can get confusing to know where
  exactly each tree is located, which can complicate things like
  collision detection and selections. In the illustration above, the
  tree is aligned to the bottom left corner of its cell, which means the
  actual location of the tree is actually in an empty part of the tile.
  Tilesets have a Drawing Offset property you can tweak to make it so
  that the base of the tree is centered on the cell, but a single offset
  may not work for every tile in the tileset.
  - This issue will be mitigated if per-tile drawing offsets are
    implemented in Tiled (see [issue
    \#871](https://github.com/mapeditor/tiled/issues/871)).

## Moving Tiled Files

If you move a Tiled Map or Tileset file relative to any files it uses,
such as tilesets or images, the references to those files will become
broken. When you open that Map or Tileset in Tiled, Tiled will be unable
to find the files and will prompt you for them. Until you fix those
references, Tiled will not be able to display any tiles, objects, or
image layers that use them.

> [!NOTE]
> In the list of missing files, you can select multiple entries at once
> and select the directory all those files are in. This can save you
> lots of time if you have a lot of missing files that are all in the
> same place.

You can avoid breaking file references by always moving Maps, Tilesets,
and the files they need together. This is much easier if you practise
good file discipline from the start:

- Make a directory (folder) for your project and keep all your related
  files inside it, either right in there or in subdirectories
- When downloading new images for your project, put them into your
  project directory before you use them in Tiled
- If you need to move your files, move the entire directory as one
- When sharing your files, keep the directory structure
  - For quick sharing, just zip up the whole project directory
  - For collaborating with others on your maps, use
    [git](https://en.wikipedia.org/wiki/Git) or similar tools. This'll
    automatically preserve all the file structure and avoid the need to
    upload/download redundant files

In addition to being a good organizational tool even if your files
aren't part of some greater project, creating a project directory like
this will also make it easier to use Tiled's Project feature and often
avoids issues when importing your Tiled files into game engines.

> [!NOTE]
> If you're using the Unity engine and using
> [SuperTiled2Unity](https://seanba.itch.io/supertiled2unity) to import
> Tiled Maps, make your Tiled project directory inside of your Unity
> project's Assets directory. This way, you will not need to move or
> copy files at all, so you'll not risk breaking any references. As a
> bonus, ST2U will be able to automatically detect changes to your files
> and reimport them.
>
> If you're using Automapping alongside ST2U, you should put your rule
> maps into a subdirectory with a name that ends with `~`. This will
> prevent Unity from trying to import your rules, avoiding a lot of
> unnecessary work, and avoiding issues with importing the Automapping
> Rules Tileset, which ST2U will not be able to find as it's built into
> Tiled and does not exist in your Unity project.

### Embedded Tilesets

It can be tempting to embed your Tilesets into your Map files to reduce
the number of tiles you have to deal with. However, this has its own
problems. As a general rule, **do not embed Tilesets**. Even if your
game engine or Tiled map parser requires embedded Tilesets, it's better
to use the "embed tilesets" export option to embed them only into your
production files, and keep your working files using external tilesets.

The main disadvantage of embedded Tilesets is that, because each copy is
separate, any changes to the tileset must be applied to every copy. If
you want to add new Terrains, for example, you'll have to add them to
the Tileset for every map where you want to use those Terrains. The same
goes for custom properties, tile animations, collisions, tile order, and
so on. Aside from the obvious hassle of ensuring every copy has all the
latest changes, it means duplicating a lot of data. If you use external
Tilesets, you only have to load each Tileset once in your game, and then
you can reuse that data for other maps without having to parse the
Tileset file again. For embedded Tilesets, you have to read and parse
the map and all its embedded Tilesets every single time, even if the
data is identical to data loaded for a previous map.

Embedded Tilesets can also cause problems for Automapping, since
Automapping needs *exact* tile matches. Automapping attempts to match
tiles from embedded tilesets, but if there are minor differences in the
Tilesets, it could think they're different Tilesets and not match tiles
that you want to match. External Tilesets are the most reliable option.

Embedded Tilesets are, however, a decent option for situations where you
know you'll only be creating a single map, and will not use those
Tilesets outside of that one map. Don't forget that Automapping rules
are additional maps, however - if you want to use that feature, you
should stick to external Tilesets even if you only have one working map.

## Parsing Tiled Maps

If you're parsing Tiled files yourself rather than using [an existing
library](https://doc.mapeditor.org/en/stable/reference/support-for-tmx-maps/)
or exporting in an engine-specific format, you may run into some of
these issues when trying to read the Tile Layer data. Fortunately, they
all have relatively simple solutions.

### Garbled-looking Tile Layer Data

Tiled supports several data formats for the list of tiles that make up
each Tile Layer. It's quite likely that you'll find layer data that
looks something like this:

    eJztzwENwCAQBEEQVKQhDWnVUBPNPYEZA5ttDQAAAAAAgD+Ny/oz3Kvur3Cvuu/37P4b7lX3Rw8Hi/tPNrddHyDlA1RLB+E=

This can be rather intimidating since it doesn't look like the list of
tile IDs you might be expecting, but don't worry. The map contains
information to help you parse this data, *and* Tiled has options for
formats that you might find easier to understand.

The Tile Layer Format can be changed in Tiled in the Map Properties,
this determines how the layer data is stored in the map files. The
following options are available:

- **XML (deprecated)**: Each tile is stored as an XML element. This
  format is offered only for backwards compatibility, and should not be
  used as it makes for needlessly large files and is slow to parse.
- **Base64 (uncompressed)**: The data is a list of tile IDs as unsigned
  32-bit integers, base64 encoded. This format has some niche uses, but
  usually one of the compressed formats is better.
- **Base64 (gzip compressed)**: The data is a list of tile IDs as
  unsigned 32-bit integers, compressed with gzip, and then base64
  encoded. This or the other compressed formats should be used whenever
  possible, as they make the data much smaller. gzip and zlip libraries
  are also commonly available either as part of engines or as
  plugins/scripts.
- **Base64 (zlib compressed)**: As above, but compressed with zlib.
- **Base64 (Zstandard compressed)**: As above, but compressed with
  Zstandard. While this is a good compression format, it's not as widely
  supported as zlib and gzip, so I don't recommend using it unless
  you're sure it's supported in your engine.
- **CSV**: The data is a list of tile IDs as CSV. This is the easiest
  format for a human to read, but it makes for larger files than the
  compressed formats. If you're just getting started and just want to
  get some tiles drawn in your game, this is a good place to start, but
  I don't recommend using CSV in production since it makes for large
  files, and can be slow to parse if you don't optimize.
  > [!NOTE]
  > In TMJ (Tiled Map JSON) maps, the "CSV" format actually outputs a
  > native JSON array rather than a CSV string. This means rather than
  > parsing the CSV yourself, you should be able to use your JSON
  > library's built in array parsing.

As mentioned above, if you're just getting started, *try changing your
map's Tile Layer Format to CSV*, as it's probably the easiest format to
work with when you're just trying to get the basics working, as you can
just look at the data and know roughly what it corresponds to, unlike
the Base64-encoded formats. When you're ready to parse the other formats
(except XML, which isn't worth bothering with), come back here.

#### Parsing the Base64-encoded formats

If you look at the data field in the layer, it should have `encoding`
and (optionally) `compression` properties. The examples here will use
the TMX format, but this is all pretty similar with JSON format and any
other export formats that respect the Tile Layer Format property in your
maps. If the `encoding` is "csv", then you're dealing with a CSV string,
which should be fairly easy to figure out. If the `encoding` is
"base64", then read on.

The Base64 formats are should all be dealt with roughly the same way:

1.  **Decode base64**. The result will be a bunch of binary data,
    typically an array of bytes, though different languages and
    frameworks have slightly different ways to deal with this.
2.  **If a `compression` is set, decompress**. The compression options
    are "zlib", "gzip", and "zstd", and you'll need to run the
    appropriate decompression algorithm based on the value of the
    property. If the `compression` property isn't set, then skip this
    step, because the data wasn't compressed. Again, the exact way to do
    this will depend on the language and libraries you're using, but
    zlib and gzip libraries are widely available. Generally, they'll
    take a byte array or a string as an input, and output a decompressed
    byte array or string.
3.  **Reinterpret the bytes as 32-bit unsigned integers**. The base64
    decoder and the decompression libraries operate at the byte level
    and don't care what the bytes mean, but what you need out of the
    data is 32-bit unsigned integers. The data is always in
    little-endian byte order, regardless of the endianness of the system
    it was written on. Most architectures these days are little-endian,
    so it's sufficient to just copy the binary data into an array of
    unsigned 32-bit integers or to reinterpret the existing data in
    memory (e.g. via casting). If you're compiling for a big-endian
    system, you'll want to rearrange the bytes (swap bytes 0 and 3 and
    bytes 1 and 2 of every set of 4 bytes) before doing that.

This might sound complicated, but once you have the decompression
libraries you need, the actual code should be fairly simple. Here's some
C++ that's very similar to what I do in my engine, using
[zlibcomplete](https://github.com/rudi-cilibrasi/zlibcomplete) for
decompressing zlib and gzip:

    //get the layer data from the file:
    std::string data = dataNode->value();

    //Prepare a container for the layer data:
    unsigned int* tileGIDs = new unsigned int[mapWidth*mapHeight];

    if(std::strcmp(encoding, "base64") == 0) {
        data = base64_decode(data);

        if(std::strcmp(compression, "zlib") == 0) {
            zlibcomplete::ZLibDecompressor decompressor;
            data = decompressor.decompress(data);
        } else if(std::strcmp(compression, "gzip") == 0) {
            zlibcomplete::GZipDecompressor decompressor;
            data = decompressor.decompress(data);
        } else if(std::strcmp(compression, "zstd") == 0) {
            //unsupported compression
            //report an error, clean up, abort loading
        }
        //Copy the bytes into the unsigned int array:
        memcpy(tileGIDs, data.data(), data.size());
    } else if(std::strcmp(encoding, "csv") == 0) {
        //parse the data string as CSV, into the tileGIDs array:
        parseCSV(data, tileGIDs);
    }

### Very Large Tile IDs

If you encounter tile IDs in the data that are much larger than the
maximum tile ID you expect, these are probably flipped/rotated tiles.
Tiled uses the most significant four bits for flip and rotation flags.
Before you can get at the tile ID, you should read these flags and store
their state so that you can apply these transformations to the tile when
you render, and then you need to clear the flags, so that only the tile
ID remains.

The [Tiled
docs](https://doc.mapeditor.org/en/stable/reference/global-tile-ids/)
have more information on these flags and parsing them, including a code
example.

### Tile IDs slightly off

If you're rendering your tiles successfully but getting the wrong tile
from what you expect, you're probably neglecting the `firstgid` of your
tileset. A Tiled Map can have tiles from multiple Tilesets, and there
needs to be some way to tell which Tileset a tile is from. Just using
the tile ID directly wouldn't work, since each Tileset has its own,
independent tile IDs, typically starting from 0. Instead of using these
conflicting IDs, every tile in a map uses a *global ID*, or a gid for
short, so that every tile can have a unique identifier even if there are
multiple Tilesets in the map.

The Tiled documentation has [a section on dealing with
`firstgid`s](https://doc.mapeditor.org/en/stable/reference/global-tile-ids/#mapping-a-gid-to-a-local-tile-id).

> [!NOTE]
> Maps that have only one Tileset have a `firstgid` of 1 for that
> Tileset. If your parser only supports single-Tileset Maps, you can
> generally assume that Tileset will always have `firstgid` 1.
>
> Although `firstgid`s are usually assigned based on the previous
> Tileset's tile count, this is *not* a guarantee and should not be
> relied upon. Tilesets' sizes may change after the map is saved, and
> Tiled could easily start basing the `firstgid` on the tiles that are
> actually used within the map instead. When dealing with multiple
> Tilesets, always read the `firstgid` from the Map instead of guessing.

### Loading Tilesets

Before you can use the tile meant by some GID, you need to load the
tileset. This is mostly a straight-forward process if you follow the
TMX/TMJ specification. It is simplest to load tilesets immediately as
you read the list of Tilesets in the Map, but it may be a good idea to
defer loading them until you know which tilesets are actually used by
the map. You should also consider keeping previously loaded Tilesets and
their images in memory to avoid re-loading them every time you load a
different map, but this sort of asset management is beyond the scope of
this tip sheet.

Image Collections have each tile defined with its own image (which may
be the same across tiles, or may be different) and the rectangular
region of that image the tile uses (defaulting to the entire image).
When loading Image Collection Tilesets, you should be able to allocate
space as specified by the Tileset's `tilecount`, and then load the
tiles. Be mindful of the fact that a Tile's ID may differ from its index
in the list of tiles - tiles may have been reordered, and tiles may have
been deleted. It's usually a good idea to use an id-Tile dictionary to
store and access Tiles in an Image Collection, rather than a plain
array, as using an array for this potentially sparse data may waste a
lot of memory.

Based on Tileset Image Tilesets often don't include explicit entries for
each tile, since each tile's position can be calculated from the tile
ID. When entries are included, it's typically because the tile contains
additional data (Class, Terrains, Custom Properties, Animations, etc),
or because the tiles have been reordered. Although the tiles may appear
out of order in the Tileset, you can use their IDs as indices into an
array if you want, since this type of tileset has no gaps in tile IDs.

> [!NOTE]
> The `tilecount` and image `width` and `height` specified in a Based on
> Tileset Image Tileset reflect the state of the image when the Tileset
> was written, and may not match the actual size of the image. That's
> harmless by itself, but it's possible that Maps use tiles outside of
> `tilecount` or numbered differently from what would be expected from
> the saved `width` and `height`, since it's possible to change the
> tileset image and use the added or renumbered tiles without saving the
> Tileset file and causing it to reflect the current image state. There
> is no way to automatically determine whether this has happened.
>
> For consistency with Tiled, it's best to derive the tile count and
> tile positions from the image's current size instead of trusting the
> values stored in the Tileset. This way, in the unlikely scenario that
> things go wrong, the problem will look the same both in-engine and in
> Tiled, making it easier to find the cause. You may also want to
> display a warning to the user when the image dimensions stored in the
> tileset do not match the image's actual dimensions. Only used the
> stored values if it is impossible to get the image's actual size.

For rendering tiles from tilesets, see below.

## Rendering Tiled Maps

Most engines support tilemaps natively, so if you're using an existing
engine, you probably don't need to care about rendering them, and only
need to worry about [getting the Tiled map data into your
engine](https://doc.mapeditor.org/en/stable/reference/support-for-tmx-maps/).
However, if you're writing your own tilemap renderer or otherwise need
some help in figuring out how tiles in Tiled connect to drawable
entities in-engine, this section may be of use.

### Drawing tiles from a Tileset

Tiled supports two kinds of Tilesets: Image Collections, and Based on
Tileset Image.

#### Image Collections

In Image Collections, every tile is its own separate image, so to render
a given tile, you'd look up its image and render it. As of Tiled 1.9,
these Image Collection tiles have `x`, `y`, `width`, and `height`
properties which determine where exactly in the image the tiles are, so
you should draw the subrectangle of each tile's image that's defined by
these properties.

#### Based on Tileset Image

Tilesets that are Based on Tileset Image are all subrectangles of the
Tileset's source image, but their locations are not explicitly defined
anywhere. Instead, they can be simply calculated from the tile ID like
this:

    x = tileID % columnCount; //in tiles
    x = x * (tileWidth + spacing) + margin; //now in pixels
    y = floor(tileID / columnCount); //in tiles
    y = y * (tileHeight + spacing) + margin //now in pixels

At the end of this, x and y contain the pixel coordinate of the top left
pixel of your tile. To get the other three corners, add tileWidth to x
and tileHeight to y as needed.

The column count is the width of the Tileset in tiles. It's included as
a `columns` field on the Tileset in most versions of Tiled, and can be
calculated for very old tilesets, which don't have it, from the image
width, tile width, margin, and spacing:

    columnCount = floor( (imageWidth - margin + spacing) / (tileWidth + spacing) );

> [!NOTE]
> Tiled currently has a single `spacing` value and a single `margin`
> value and these unified values are used in the examples above, but
> it's possible that one day, it will be possible to set these values
> independently for x and y. I recommend using separate x and y
> variables for each in your calculations, and just assigning the
> current unified value to both for now, so that in the future, you only
> have to update your code that reads these values, and not your
> calculations.

### Rendering Maps

While it is outside the scope of this document to describe 2D video game
rendering, this section will cover where each grid cell is positioned
for the different map orientations in Tiled, as this is not always
trivial, and few engines natively support non-orthogonal orientations,
and even those that support orthogonal orientations don't always support
drawing the tiles in all the ways Tiled does.

In general, you should render layers from bottom to top, i.e. in the
order that the layers appear in the Tiled map file, and each layer is
largely independent. Each layer in Tiled can have its own `offsetx` and
`offsety`, this means the layer should be drawn that many pixels right
and down from its default location; in most engines you can simply
subtract the offset from the layer's origin. If that's not an option,
you'll need to add the offset to the layer's render position every time
you draw it.

> [!NOTE]
> For simplicity, this section speaks of "drawing" tiles in various
> orders and locations. You should not actually render tiles one by one,
> that's very slow! Instead, add them to a single drawable entity
> (vertex array, sprite batch, etc) in that order and at that location,
> and draw that entity all at once.

Tiled supports four map orientations: orthogonal, isometric, isometric
(staggered), and hexagonal. The actual tile textures are always
rectangles, and the rendered layer can be thought of as a screenspace
rectangle, so all these orientations change is where each cell is. The
cell is the slot where a tile can be, the cells the spaces between the
grid lines that you can see in Tiled. **Cells and tiles aren't the same
thing.** A cell may be empty, which means no tile should be drawn there,
and more importantly, tiles may stick out of their cells, may not fill
the entirety of the cell, and may be offset from their cells - sometimes
all at the same time. The rest of this section will discuss cell
positions, but to figure out where to draw the tile relative to the
cell, you'll need to do some additional work, in this order:

1.  Calculate the tile's render size and scale. By default, this is the
    same as the tile size, so both the x and y scale factors are `1.0`,
    but `tilerendersize` on the Tileset may be set to `grid`, which
    means the tiles should be scaled to fit the grid size. In that case,
    the exact scaling depends on the `fillmode`. If it's `stretch`
    (default), then divide the cell width by the tile width to obtain
    the x scale factor, and the cell height by the tile height to obtain
    the y scale factor. If it's `preserve-aspect-fit`, then do the same,
    but then use the smaller value of the two for scaling both the width
    and height of the tile. The tile's render size is its base size
    multiplied by the calculated scale factor. Since tiles can differ in
    size and shape, this calculation has to be done for each tile,
    though you can cache the results since within a map, each cell is
    the same size.
    > [!NOTE]
    > In all Maps, the cells are the map's `tileWidth` in width and the
    > map's `tileHeight` in height. The different orientations only
    > change *where* each cell is, not the cells' size.

2.  To match Tiled, tiles should be drawn with bottom left alignment,
    that is, the bottom left corner of the tile's bounding box should be
    in the bottom left corner of the cell's bounding box. If the tile is
    drawn larger than the cell, that means it'll stick out at the top
    and left side, and if the tile is drawn smaller than the cell, it'll
    leave space above and to the right. Since most rendering happens
    relative the top left, that means you'll need to subtract the tile's
    render height from the bottom left corner position, this'll give you
    the location of the top left corner of the tile for that cell. This
    has to be done for each tile independently, as tiles may have
    differing sizes.

    If the `fillmode` is `preserve-aspect-fit`, the tile should be
    centered within its bounding box. Add half the difference between
    the cell size and tile's scaled size to the render position. For
    simplicity, you can apply this to `stretch` Tilesets too - the
    difference will just always be `0, 0` and not change the render
    position.

3.  Tilesets have a `tileoffset` property, this should be scaled by the
    render scale and then added to the tile's render position.

This process is the same regardless of map orientation, so in the
orientation-specific sections below, I will focus on calculating the
bottom left origin for each tile. From this, you can calculate the
position at which to render the tile following the process above.

> [!NOTE]
> If this looks intimidating, then start by ignoring the
> `tilerendersize` - assume tiles should be rendered at their native
> size. This was how Tiled worked until 1.9, and it's how most tilesets
> are designed to work. You can add the scaling later if you need it.

In the orientation-specific sections below, `x` and `y` are always the
cell's coordinates, i.e. in map-space, and are assumed to be integers.
`gridWidth` and `gridHeight` are the map's tile size, which may be
different from the size of the tiles. These pseudocode examples are
based on the code Tiled internally uses to translate map-space
coordinates to screenspace coordinates.

#### Orthogonal Map cells

Orthogonal Maps are the simplest. The four sides of a cell's bounding
box are simply:

    left = x * gridWidth;
    right = left + gridWidth;
    top = y * gridHeight;
    bottom = top + gridHeight;

The bounding box of an orthogonal cell is the cell itself. This is not
true for the other orientations.

For orthogonal maps only, Tiled supports the Tile Render Order property,
which determines the order in which tiles are drawn, affecting how tiles
that stick out of their cells overlap. All this really affects is the
direction you iterate x and y in when drawing the tiles, so that tiles
drawn first are on the bottom, overlapped by tiles drawn later. The
render orders starting with "Right" go from left to right, meaning x
starts at `0` and increases. Those starting with "Left" are the
opposite, they start at `mapWidth - 1` and decrease. Render orders
ending in "Down" go from top to bottom, meaning y starts at `0` and
increases, while those ending in "Up" start at the bottom, so y starts
at `mapHeight -1` and decreases.

#### Isometric Map cells

Isometric Maps are diamond shapes with 0,0 at the top corner, x
increases towards the lower right, y increases towards the lower left.
You can calculate the bounding rect of the isometric cell thus:

    originX = mapHeight * gridWidth / 2;
    xPixel = (x - y) * gridWidth / 2 + originX;
    yPixel = (x + y) * gridHeight / 2;
    //xPixel and yPixel are at the upper corner of the cell,
    //i.e. the middle of the top edge of the bounding box.
    left = xPixel - gridWidth / 2;
    right = xPixel + gridWidth / 2;
    top = yPixel;
    bottom = yPixel + gridHeight;

`originX` is the x coordinate of the origin point in the map. Every row
of tiles adds a tile's worth of width to the map's screenspace width, so
the map's width is `mapHeight*gridWidth`, and `originX` is half that
since the map's origin is in the horizontal middle of the map in
screenspace.

The render order for isometric maps is back to front. Since these maps
are skewed, this means you can't just iterate each row, the ordering is
more complex. A simple method to iterate the cells this way is to
iterate the sum of x and y:

    for(sum = 0; sum < mapWidth + mapHeight - 2; ++sum) {
        for(x = 0; x < sum && x < mapWidth; ++x) {
            y = sum - x;
            if(y < mapHeight)
                //render the tile at x, y
        }
    }

This method is convenient if you're building a single drawable entity
for the whole layer. If you need to limit which parts you draw based on
some screenspace rectangle, such as when using software rendering, then
it may be more beneficial to iterate in screenspace, stepping by
`gridWidth` in x and `gridHeight / 2` in y, convert each location to map
space, and if it's the coordinates are within the map, draw that cell.
Screenspace to map-space conversions are outside the scope of this text,
but you can take a look at `IsometricRenderer::drawTileLayer` and
`IsometricRenderer::screenToTileCoords` in Tiled's source code for
inspiration.

#### Staggered Map cells (Hexagonal and Isometric)

Staggered maps are roughly rectangular in shape, but have isometric or
hexagonal cells. Isometric staggered maps are practically identical to
hexagonal staggered maps, except that their `sideLength` is always `0`.
The stagger axis has a large effect on how coordinates are calculated,
and the stagger index also comes into play. The calculations for these
maps are more complex because of the staggering, some cells are offset
relative to their neighbours.

    //Some useful functions and values:
    bool shouldStaggerX(int x) {
        if(staggerX) {
            if(x % 2 == 0) //x is even
                return staggerEven;
            else
                return !staggerEven;
        }
        return false;
    }
    bool shouldStaggerY(int y) {
        if(!staggerX) { //if staggerY
            if(y % 2 == 0) //y is even
                return staggerEven;
            else
                return !staggerEven;
        }
        return false;
    }

    columnWidth = gridWidth / 2;
    rowHeight = gridHeight / 2;
    if(staggerX)
        columnWidth += sidelength / 2;
    else //staggerY
        rowHeight += sideLength / 2;

    //Finally, the actual cell calculations:
    if(staggerX) {
        xPixel = x * columnWidth;
        yPixel = y * (gridHeight + sideLength);
        if( shouldStaggerX(x) )
            yPixel += rowHeight;
    } else { //staggerY
        xPixel = x * (gridWidth + sideLength);
        if( shouldStaggerY(y) )
            xPixel += columnWidth;
        yPixel = y * rowHeight;
    }
    left = xPixel;
    right = xPixel + gridWidth;
    top = yPixel;
    bottom = yPixel + gridHeight;

`staggerX` is a boolean indicating whether the map's stagger axis is X
(`true`) or Y (`false`), and `staggerEven` is a boolean that determines
whether the stagger index is even (`true`) or odd (`false`).
`columnWidth` and `rowHeight` differ from `gridWidth` and `gridHeight`
because the rows or columns are staggered, meaning they overlap.
Fortunately, these two values are consistent across the entire map, so
you only need to compute them once. The `shouldStaggerX` and
`shouldStaggerY` functions check whether the current tile should be
staggered, their output depends on the map's stagger index and the
tile's position. In languages where booleans can be coerced into
integers, these two functions can be rewritten as one-liners that avoid
branching logic:

    bool shouldStaggerX(int x) {
        return staggerX && (x & 1) ^ staggerEven;
    }
    bool shouldStaggerY(int y) {
        return !staggerX && (y & 1) ^ staggerEven;
    }

> [!NOTE]
> In this example code, I split the staggerX and staggerY functions. You
> can inline the stagger checks too, but it's easier to read when
> they're separated like this. If your language supports it, you can
> make the functions `inline`, to get the readability benefits of
> separate functions without the performance downside.

Like isometric maps, staggered maps are rendered back to front, and the
exact way to achieve that depends on the stagger axis. If it's Y, then
just iterate by rows:

    if(!staggerX) {
        for(y = 0; y < mapHeight; ++y) {
            for(x = 0; x < mapWidth; ++x) {
                //render the tile at x, y
            }
        }
    }

If the stagger axis is X, then for each row, you'll need to first draw
every odd tile (if stagger axis is even) or every even tile (if stagged
axis is odd), and then draw the remaining tiles (even or odd):

    else {
        for(y = 0; y < mapHeight; ++y) {
            startX = 0;
            if(staggerEven)
                startX = 1;
            for(x = startX; x < mapWidth; x += 2) {
                //render the tile at x, y
            }
            
            if(startX == 0)
                startX = 1;
            else
                startX = 0;
            for(x = startX; x < mapWidth; x += 2) {
                //render the tile at x, y
            }
        }
    }

## Scripting

Tiled supports scripting via JavaScript and via Python, but the Python
API is out of date and it's not recommended to use it. This section will
deal only with the JavaScript API. Details on the JavaScript scripting
system can be found in the [Tiled scripting
documentation](https://doc.mapeditor.org/en/stable/reference/scripting/).
This section will attempt to fill some gaps and inconveniences of the
official docs, which unfortunately have a poor search feature, making it
difficult to find things if you don't already know what you're looking
for. As such, there will be few details in this section, instead there
will there will be many links to the API documentation.

### CLI

Scripts are typically run via the Tiled Editor GUI, but they can also be
run via the CLI: `--export-map` and `--export-tileset` can use scripted
Map and Tileset formats, and you can execute arbitrary scripts with
`--evaluate <scriptFile> [args]`. The latter allows you to pass
additional parameters to the script as well, which you can access in the
script via
[`tiled.scriptArguments`](https://www.mapeditor.org/docs/scripting/modules/tiled.html#scriptArguments).

When running scripts via CLI, none of the GUI features are available.
This includes most of the functionality in the Tiled GUI section below,
methods like `tiled.trigger()` and `tiled.open()`, and, perhaps
surprisingly, `TileMap.automap()`, as Automapping is part of the Tiled
Editor GUI and not part of the core Tiled library.

If you want to write scripts that work both via CLI and GUI, you should
avoid relying on the GUI-specific features and `tiled.scriptArguments`.
If your script needs the current document, the way to do this will vary:
if the user is running the GUI, then `tiled.activeAsset` will work, and
if that's `null`, you can check `tiled.scriptArguments` for a path,
which you can then read.

### GUI

When running scripts in the Tiled GUI, they can access many parts of the
GUI, so your scripts can respond to the user's current brush, chosen
tileset, current document tab, etc. These parts of the API are not
available to scripts running via the CLI.

You can get the current active document via
[`tiled.activeAsset`](https://www.mapeditor.org/docs/scripting/modules/tiled.html#activeAsset),
and the list of all open documents via
[`tiled.openAssets`](https://www.mapeditor.org/docs/scripting/modules/tiled.html#openAssets).

Tiled's GUI consists of two major parts, each with their own selection
of panels: the Map Editor, which is shown when the active document is a
Map, and the Tileset Editor, which is shown when the active document is
a Tileset. Understanding this split is important if you want to write
scripts that interact with the GUI. For example, the Terrain chosen in
the Map Editor's Terrain Sets panel is used for painting on Maps and has
nothing to do with the Terrain chosen in the Tileset Editor's Terrain
Sets panel, which is used for assigning Terrains to Tiles.

#### Map Editor

Scripts can interact with the following Map Editor features:

- [the current
  brush](https://www.mapeditor.org/docs/scripting/interfaces/MapEditor.html#currentBrush)
  (`tiled.mapEditor.currentBrush`)
- [the Map
  View](https://www.mapeditor.org/docs/scripting/interfaces/MapEditor.html#currentMapView)
  (`tiled.mapEditor.currentMapView`), which lets you change the zoom and
  focus of the view for the current map
- [the Tilesets
  View](https://www.mapeditor.org/docs/scripting/interfaces/MapEditor.html#tilesetsView)
  (aka Tilesets panel; `tiled.mapEditor.tilesetsView`), which lets you
  change which tileset is displayed, and which tiles in it are selected
- indirectly, the Terrain Sets panel, you can get (but not set) the [the
  current Terrain
  Set](https://www.mapeditor.org/docs/scripting/interfaces/MapEditor.html#currentWangSet)
  (`tiled.mapEditor.currentWangSet`) and [current
  Terrain](https://www.mapeditor.org/docs/scripting/interfaces/MapEditor.html#currentWangColorIndex)
  (`tiled.mapEditor.currentWangColorIndex`)

The Layers and Objects panels aren't accessible via scripts directly,
but they reflect the state of the current Map, so you can interact with
them through the Map document:
[TileMap.layers](https://www.mapeditor.org/docs/scripting/classes/TileMap.html#layers),
[ObjectGroup.objects](https://www.mapeditor.org/docs/scripting/classes/ObjectGroup.html#objects)
(Objects are accessed through their Object Layers)
[TileMap.selectedLayers](https://www.mapeditor.org/docs/scripting/classes/TileMap.html#selectedLayers),
[TileMap.selectedObjects](https://www.mapeditor.org/docs/scripting/classes/TileMap.html#selectedObjects).

#### Tileset Editor

Scripts can interact with the following Tileset Editor features:

- indirectly, the Terrain Sets panel, you can get (but not set) the [the
  current Terrain
  Set](https://www.mapeditor.org/docs/scripting/interfaces/TilesetEditor.html#currentWangSet)
  (`tiled.tilesetEditor.currentWangSet`) and [current
  Terrain](https://www.mapeditor.org/docs/scripting/interfaces/TilesetEditor.html#currentWangColorIndex)
  (`tiled.tilesetEditor.currentWangColorIndex`)
- [the Tile Collision
  Editor](https://www.mapeditor.org/docs/scripting/interfaces/TilesetEditor.html#collisionEditor)
  (`tiled.tilesetEditor.collisionEditor`), where you can control the
  selected collision objects and modify the view similar to the map
  view. To actually modify the collision objects of a Tile, you should
  go through the [Tile
  itself](https://www.mapeditor.org/docs/scripting/classes/Tile.html#objectGroup).

### Scripting Caveats

There are a number of quirks to the scripting API - some due to bugs,
some due to the API functions originally being written for purposes
other than scripting, some due to Qt/Tiled limitations. This is not an
exhaustive list.

When getting a property from a Tiled entity, such as
`TileMap.selectedLayers` or `tiled.mapEditor.currentBrush`, you get a
copy, your changes to it will not affect the Tiled/Qt side. That is,
most things are returned by value, not by reference. You need to assign
back to that property if you want to change the value. The exceptions
are properties that are references by definition, such as
`tiled.activeAsset`.

When editing Tile Layers with `setTile()`, the `apply()` function
currently takes the user's selection into account, so any modifications
you make that are outside the user's selection will not be applied. You
can get around this by clearing the selection, e.g.
`map.selectedArea = Qt.rect(0,0,0,0)`. You should do this any time that
you're modifying a document the user may have previously interacted
with, unless you specifically want to only modify tiles within their
selection (in which case, you can make your script perform better by
only doing the work within the selection in the first place). This is
[issue \#3482](https://github.com/mapeditor/tiled/issues/3482).

- If you clear the user's selection to avoid this issue, consider saving
  it to a variable first, and then restoring it.

Although it is a property, `Tileset.tiles` creates JavaScript-side Tile
objects every time you access it, making it rather slow. Instead of
accessing it repeatedly, save it to a local variable, and then do work
with that variable, as much as you can.

Resizing a layer via `resize()`, whether directly or via resizing the
map it's in, causes the layer to be replaced with a clone, invalidating
any existing references you may have to that layer. This means that you
should avoid resizing any layers that aren't part of a map (or you won't
be able to recover them!), and if you need to compile a list of layer
references for a map, you will want to do it *after* any resizing. This
is tracked as [issue
\#3480](https://github.com/mapeditor/tiled/issues/3480). You can resize
by changing a layer's or map's `width` and `height` properties without
issue.

There is currently no way to open a Tiled document "silently" - without
displaying it to the user. This becomes an issue when you need to open
Tilesets to add to a TileMap in a custom map format, for example. It's
possible to read the contents of a Map or Tileset file with the
appropriate format's `read()` method, but this creates a copy
disconnected from the original file, and using a Tileset loaded this way
in a Map would create an embedded Tileset. There might eventually be
[something like
tiled.load()](https://github.com/mapeditor/tiled/issues/3517#issuecomment-1311716350)
that will open a document without displaying it to the user.

`tiled.open()` is the correct way to go in this case, at least for now.
You can `tiled.close()` the document after you're done with it, though
you'll want to avoid closing any documents that were already opened by
the user. You can achieve this by first checking `tiled.openAssets` for
the document you want, and only doing `open()` and `close()` if you
don't find it there.

All Tiled scripts with the .js extension share the same global context.
Any persistent variables you declare in one script can be accessed in
any code that runs later, even from other scripts. You should avoid
creating more global variables than you need, and you should be careful
to name them specifically enough that they're not likely to collide with
the variables of another script. Alternatively, give your script the
.mjs extension, so it will be loaded as a module with its own context.
It should still work otherwise the same, even if it isn't actually a
module that imports/exports any functionality.

This isn't a quirk, but rather a natural consequence of having
JavaScript-side proxy objects for objects on the C++ side: you can
assign any property on the JS objects, but only properties that are
valid for the C++ objects will make it to the C++ side for use and
display in Tiled. You can, however, take advantage of this to
conveniently store temporary data about Tiled objects. For example, if
you register an Action in Tiled, you can store its configuration on the
Action object in JS, without interfering with the action's
functionality, without filling up the global space with a bunch of
variables.

When an Asset is reloaded (via Ctrl+R or because the file is changed
externally; this can only happen to Assets that are open as tabs in the
Tiled GUI), your Asset object will refer to the newly reloaded data, but
any stored references to its internals (layers, etc) will refer to
stand-alone copies of the old data. If you're storing such references,
you should listen for `tiled.assetReloaded` and update those references.
You may also want to listen to `tiled.assetAboutToBeClosed` to know to
clear your references to the Asset and any of its components.

### Reading and Writing Files

#### Maps and Tilesets

If you want to open a Map or Tileset in Tiled that isn't already
supported, you should use
[`tiled.registerMapFormat()`](https://www.mapeditor.org/docs/scripting/modules/tiled.html#registerMapFormat)
and
[`tiled.registerTilesetFormat()`](https://www.mapeditor.org/docs/scripting/modules/tiled.html#registerTilesetFormat),
which will allow many parts of Tiled and other scripts to work with this
format. Formats let you define `read()` and `write()` methods, but it's
fine to only define one of them if you don't need the other. If your
format only supports `write()`, you'll only be able to save to it via
Export As, not Save As.

If you need to write a Map or Tileset to a file in a format Tiled
supports (whether by default, or with a custom format), you can use the
appropriate
[MapFormat](https://www.mapeditor.org/docs/scripting/interfaces/MapFormat.html)'s
or
[TilesetFormat](https://www.mapeditor.org/docs/scripting/interfaces/TilesetFormat.html)'s
`write()` method. You can get the Format for your desired file type with
[`tiled.mapFormat(shortname)`](https://www.mapeditor.org/docs/scripting/modules/tiled.html#mapFormat)
and
[`tiled.tilesetFormat(shortname)`](https://www.mapeditor.org/docs/scripting/modules/tiled.html#tilesetFormat).
You can also get lists of each type of format via
[`tiled.mapFormats`](https://www.mapeditor.org/docs/scripting/modules/tiled.html#mapFormats)
and
[`tiled.tilesetFormats`](https://www.mapeditor.org/docs/scripting/modules/tiled.html#tilesetFormats).

If you need to read a Map or Tileset file in a format that Tiled
supports, you can use the MapFormat and TilesetFormat as above, but
using `read()` instead of `write()`. This will create a new TileMap or
Tileset object for you to use, but will *not* open it in the GUI editor,
so it will not have Undo states or any other GUI-related features
available, and it'll be a brand new Asset disconnected from its original
file, with no `fileName`. You can open it as a document in the GUI by
assigning it to
[`tiled.activeAsset`](https://www.mapeditor.org/docs/scripting/modules/tiled.html#activeAsset),
but it'll still be disconnected from its original file location. If you
want to open a Map or Tileset in the GUI, use
[`tiled.open()`](https://www.mapeditor.org/docs/scripting/modules/tiled.html#open)
instead, which will open it as the active document in Tiled. If you
don't want that, you can either wait for [issue
\#3517](https://github.com/mapeditor/tiled/issues/3517) to be resolved
and a `tiled.load()` method to be added, or you can save a reference to
`tiled.activeAsset` to remember what the user had open, and set it back
to that document after using `tiled.open()`. You can also
[`tiled.close()`](https://www.mapeditor.org/docs/scripting/modules/tiled.html#close)
the Asset afterwards. For a better user experience, you should only
close the Asset if it wasn't already opened by the user, which you can
check by checking whether
[`tiled.openAssets`](https://www.mapeditor.org/docs/scripting/modules/tiled.html#openAssets)
contains the Asset you need prior to opening it.

#### Other Files

If you need to read or write an arbitrary file as part of your script,
Tiled provides
[TextFile](https://www.mapeditor.org/docs/scripting/classes/TextFile.html)
and
[BinaryFile](https://www.mapeditor.org/docs/scripting/classes/BinaryFile.html)
to help. To use them, create a `new TextFile(path, mode)` (or
`new BinaryFile(path, mode)`). `path` is the location of the file to
read or write, `mode` is one of `ReadOnly`, `WriteOnly`, `ReadWrite`
(the preceding options are available for both types of file, e.g.
`BinaryFile.ReadOnly`), or `TextFile.Append`. These helper classes
provide numerous methods to help you in working with files, which you
can find in their documentation. If you're making a custom Map or
Tileset format, you'll probably use one of these as part of your reading
and/or writing code.

When writing to a BinaryFile, you'll need to use
[ArrayBuffers](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/ArrayBuffer),
you cannot write strings, numbers, etc directly. I recommend writing
some helper functions to write individual values, which create an
appropriately-sized array buffer, create a relevant DataView for it,
write the value, and then write out that buffer, e.g.

    function writeUInt32(number) {
        const buff = new ArrayBuffer(4); //size of UInt32: 4 bytes
        const view = new DataView(buff);
        view.setUInt32(0, number); //write the number at the start of the buffer
        myFile.write(buff);
    }

### Sharing Your Scripts

If a script you've written might be useful to other people, please
consider sharing it! The [tiled-extensions
repo](https://github.com/mapeditor/tiled-extensions) accepts pull
requests - you don't need to set up a git repo for your scripts to
submit them, you can submit files via the GitHub website. Before you
submit your script, consider its usability - you know your script well,
but other people will not. Some things to consider:

- If you write any temporary files, delete them after you're done. If
  you're storing user preferences in files to save them across
  executions, consider if you *really* need this - instead of saving
  configuration to files, you can store it in [custom properties on the
  user's
  Project](https://www.mapeditor.org/docs/scripting/classes/Project.html#setProperty),
  and you should consider making this configuration optional by letting
  the user edit some variables somewhere near the top of your script. If
  you must use files for configuration, read and write them as little as
  possible.

- If your script has a potentially long execution time, or if it is an
  action modifies files or documents without the possibility of Undo,
  consider adding a confirmation pop-up (at least when not running via
  CLI). Accidental executions can and do happen, especially if your
  script is an action added to a menu or toolbar.

- Try to leave the GUI as you found it, except for the changes your user
  expects the script to make. For example, if you're opening some
  documents that the user didn't ask for (e.g. tilesets for a new map),
  close them, and if you need to change the user's selection, you may
  want to restore it after you're done.

- Make icons for scripted tools. This can just be a small PNG file, you
  don't have to make a full vector icon if you don't want to. Icons
  usually take up less space than the tool name. You can also reuse any
  of the images [from Tiled's resource
  directory](https://github.com/mapeditor/tiled/tree/master/src/tiled/resources)
  by setting the icon to the relative path to them from /resources/ and
  prefixing it with `:`, e.g. `icon: ":images/16/remove.png"`.

- If your script adds an action that doesn't need to be used frequently,
  consider *not* adding it to any menus, to avoid cluttering them. As
  long as the action is named appropriately, users can find it via
  Search Actions.

- Add explanatory tooltips to your Dialog widgets.

- Since English is the default language of Tiled and most Tiled users
  have some knowledge of it, try to write your user-facing texts in
  English, unless your extension is meant speficially for users of
  another language. If you want to take it a step further, English
  comments and English function and variable names can make it easier
  for more people to troubleshoot your script.

- Add a comment at the top of the script that explains what your script
  does, where to find any actions it adds, and who wrote it (e.g. GitHub
  username). This way people browsing their script files will know what
  the file is, and where to go if they need help with it. If your script
  consists of multiple files, each file should mention which extension
  it's a part of, and ideally list all the files, so that users can
  remove them all easily once they no longer need them.

## Credits

Some examples on this page use art from the following asset packs:

- [Modified Isometric 64x64 Outside
  Tileset](https://opengameart.org/content/modified-isometric-64x64-outside-tileset)
  by Yar and darkrose
- [Kings and Pigs asset
  pack](https://opengameart.org/content/kings-and-pigs) by Pixel Frog
- [Swamp 2D
  Tileset](https://opengameart.org/content/swamp-2d-tileset-pixel-art)
  by CraftPix.net 2D Game Assets
- [LPC: Modified base
  tiles](https://opengameart.org/content/lpc-modified-base-tiles) by
  Sharm
- [Outdoor tiles,
  again](https://opengameart.org/content/outdoor-tiles-again) by
  [Michele "Buch" Bucelli](https://opengameart.org/users/buch)

This page was written by
<a href="http://eishiya.com" rel="author">eishiya</a> and posted at
<http://eishiya.com/articles/tiled/>.

