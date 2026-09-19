## Context

See proposal.md — Why. What matters here is what was already standing before this change:

- The menu draws in three passes: the scene into a 640x360 texture, that texture blown up into a 1280x720 frame with the logo and the widgets over it, and the whole frame through the palette shader into the virtual screen. Every pixel the menu shows is therefore snapped to Endesga 64, whatever it was drawn as.
- The scene is painted only from a list of the palette's reds, and a test holds every scene colour to a hue within twenty degrees of pure red.
- The menu lays out in window pixels, not in the level's 160x90 grid. The widget column is six rows of 40 pixels at x=190, with a label column 170 pixels to its left, and a test already held them inside the left 560 pixels of the window.
- The widgets were raygui's, in the level editor's dark style, loaded once at start-up.

So the frame had a palette guarantee and a red rule to obey, a layout to fit inside, and one global widget style shared with the editor. Those four are the constraints every decision below answers to.

## Goals / Non-Goals

**Goals:**

- The chrome is painted from the same list of reds as the scene, so the red rule covers the frame as well as the picture and a test can say so.
- One place decides which row is lit, whether the mouse or the keyboard moved it.
- The menu's widget style cannot leak into the level or the editor.
- The layout stays testable without a window: everything the column draws is provably inside the panel.

**Non-Goals:**

- Replacing raygui. The fields, spinners, slider, toggles and selector stay raygui's; only the buttons become rows.
- A widget framework, a theme file or a layout engine. The rows are drawn, and the geometry is constants.
- Reworking the scene's composition or the models' placement.

## Decisions

### The chrome's colours are a second block of the same red list

The chrome — rules, panel, rows, readouts — takes its colours from `menuRedIndexes`, the list the scene is painted from, and is declared beside the scene's colours in `menu_scene.go` rather than in the drawing code. The existing red-set test then covers the chrome for free by naming the new variables, and there is one place to look when the menu drifts off-red.

Alternative: free RGB for the chrome, on the grounds that the palette pass snaps it anyway. Rejected — snapping is not choosing. A hand-picked `#ff3355` lands on whichever of the 64 is nearest, which is how a menu ends up with a colour nobody decided on.

### The fill ladder runs dark to light, panel to lit row

```
  scene   sky/ground   17  (57,31,33)   darkest
  panel                17  at 0.90 alpha, over the scene
  row                  63  (87,28,39)
  lit row              62  (137,30,43)
  border, lit          60  (234,50,60)
  marker/pointer        0  (255,0,64)
```

Each layer is a step lighter than the one under it, so a row reads as raised out of the panel. The first attempt had the panel lighter than the rows, which inverted that and made the rows read as holes punched in a slab. The reds available below 17 are the palette's purples, so 17 is the floor; the panel gets its depth from alpha over the scene rather than from a darker colour.

### The CRT's gaps are the darkest red, not black

Black gaps are what a CRT actually does, and they are what the example does. Here they would blend to the palette's neutral greys, which is exactly the drift the red rule exists to prevent. The gaps are painted in the scene's darkest red at half alpha instead: they cut the grid, the sun and the rain — everything lighter than the floor — and leave the flat sky untouched. The effect reads as a CRT because the bright things are what get cut, and the scene stays inside its reds.

### The highlight is decided before anything is drawn

Input runs once per frame, before the frame is composed, and leaves two package variables: which row is lit and whether it was taken. The row drawing reads them. Moving the mouse over a row moves the highlight to it, which is why there is one answer to "what is lit" rather than a keyboard answer and a mouse answer that can disagree.

Alternative: let each row decide its own state from hover, as raygui does. Rejected — keyboard navigation has no hover to read, and a mouse resting on row three while the keyboard sits on row one would light two rows.

Only the main screen reads navigation keys. The other screens have text fields and spinners on their rows, and W or S there is a character, not a move. Their rows light on hover instead, which needs no arbitration.

### Escape goes through the same exit as the Back row

Both call one function, so a screen cannot back out one way and save its settings only the other way. It also clears the fields' edit flags, because a screen left mid-edit comes back with a field still blinking.

### raygui is restyled through DEFAULT, and the editor's style is restored on entering a level

raygui's `DEFAULT` control propagates a base property to every control, so eleven colour properties set once per frame cover the fields, the spinners, the slider, the toggles and the selector. The style is global state shared with the level editor, so `enterLevel` reloads the dark style — which is what `leaveMatch` already did in the other direction, so the pattern is the repo's, not a new one.

Alternative: a second `.rgs` style file for the menu, loaded on entering the menu. Rejected for now — it is another asset to keep in step with the palette by hand, and the eleven properties are the whole difference.

### The rain's weight, as tuned

The first pass was too faint to read in play: the scanline gaps cut most of each streak, and a shower that looked present in a still frame disappeared in motion. The values it settled at:

```
  streaks at once        160        (was 90)
  body alpha            0.65        (was 0.45)
  streak length       12..52 scene px  (was 10..46)
  tip                      3 px, in a brighter red than the body
                                    (was 2 px in the body's own red)
```

The tip is its own declared colour rather than the body's, so a streak has a lit head and a dim trail — which is what makes it read as falling rather than as a static dash. It is a palette red like everything else, so the scene's red rule still covers it.

The sweeping band was left where it was. Nothing about it read wrong; the rain was the whole of the problem.

### The panel's geometry is constants, and a test holds the column inside it

The panel is a rectangle between the two rules, inset from the left edge, and its bounds are constants next to the row geometry. A windowless test checks that the header, the logo at the scale the logo picks, all six rows with their labels and the error line are inside it, and that the panel is inside the rules. That is the same trick the layout test already played on the 560-pixel column, extended to the thing that is now actually drawn.

### The frame is chrome, and the spec says which is which

The rules cross the model area; the requirement that protects the models forbade "the logo, a button, text field, spinner, label or the error message". Rather than leave the rules as an undocumented exception, the requirement now draws the line explicitly: what you read and click stays off the models, and the picture's own chrome may cross them. The readouts are in the corners, which is where the models are not.

Alternatives weighed: inset the rules to the panel's width, which keeps the old rule intact and stops the frame from framing anything; or lift the models clear, which touches a composition another requirement pins down and risks the scenario that says all five stay inside the window.

## Risks / Trade-offs

- **The frame crosses the wireframes, and that is a matter of taste.** → It is one pixel high at the window's edges, and a capture is cheap: the rules are two constants and can be inset to the panel without touching anything else.
- **The rain and the sweeping band are tuned by eye and easy to overdo.** → Alpha and speed are single constants, and the rain is deterministic, so two captures of the same moment are comparable.
- **The small readouts use the palette's most orange red.** → It passes the twenty-degree rule, but it is the one colour in the menu that does not read as crimson. Swapping it for a deeper red is a one-line change if it grates.
- **The widget style is global state set every frame.** → Cheap, but it means the menu is the last writer while it is up. The restore on entering a level is what keeps that from reaching the editor, and it is tested by opening the editor after a match rather than by a unit test.
- **Two ways out of a screen.** → Both go through one exit, so they cannot diverge; the cost is that the exit has to know about saving settings, which is what the Back row knew before.

## Migration Plan

Nothing to migrate. No asset, save file or protocol changes: `settings.json`, the level and the network protocol are untouched. Reverting is reverting the three source files.
