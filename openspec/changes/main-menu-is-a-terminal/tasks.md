The code for groups 1 to 5 has shipped: it was written and verified in one
session, and those tasks are marked for what was actually checked. What is left
open in group 6 is what a capture cannot answer: the keys, the editor's style
after a match, and the run of the release build.

## 1. The chrome's colours

- [x] 1.1 Declare the chrome's colours — the rules, the panel, the row fills, the borders, the readouts and the markers — beside the scene's in `menu_scene.go`, every one of them an entry of `menuRedIndexes`, in a ladder that runs dark from the panel up to the lit row. Verify by naming them in the scene's red-set test and confirming `go test ./` passes.
- [x] 1.2 Run the fill ladder past a capture: the panel darker than a row, a row darker than the lit one. Verify by capturing the main screen and checking the rows read as raised out of the panel rather than cut into it (the first ladder was inverted and did read as holes).

## 2. Weather and the CRT

- [x] 2.1 Add the sky gradient above the horizon and the horizon line the grid converges on. Verify in a capture that the sky lifts toward the top of the window and the horizon is a single lit line.
- [x] 2.2 Add the rain: streaks in normalized scene coordinates from a fixed seed, a dim body with a bright tip, coming back above the top in a new column once past the bottom. Verify with a windowless test that every streak is inside the scene, none longer than a quarter of its height, each with a speed of its own, and that the same seed gives the same shower.
- [x] 2.3 Add the CRT: a band sweeping down the scene, and scanline gaps painted in the scene's darkest red rather than black. Verify in a capture that the gaps cut the grid, the sun and the rain, leave the flat sky as it is, and that the scene is still red throughout.

## 3. The frame and the panel

- [x] 3.1 Draw the two rules with their readouts: the terminal's name and connection along the top, the status that follows the session, the key hints that follow the screen, and the protocol version. Verify by capturing the menu with no session and confirming each readout is in its corner and clear of the models.
- [x] 3.2 Draw the panel between the rules with its header — a rule, the terminal's place, what it is for — and the name of the current screen under the logo. Verify by capturing the Main, Host Game and Options screens and confirming the header sits above the logo and the screen names itself each time.
- [x] 3.3 Give the logo an offset ghost of itself, keeping its fade-in timing. Verify in a capture that the ghost is behind the logo and that nothing of the logo is clipped by the panel.
- [x] 3.4 Pin the panel's geometry in constants and hold the column inside it with a windowless test: the header, the logo at the scale `logoScale` picks, all six rows with their labels and the error line inside the panel, and the panel inside the rules. Verify `go test ./` passes with the existing layout test unchanged.

## 4. Rows instead of buttons

- [x] 4.1 Draw a menu row by hand — number, marker, border, fill, and for the lit one a brighter border and fill, a pulsing glow and a pointer. Verify by capturing the main screen and confirming exactly one row is lit.
- [x] 4.2 Replace every `rlg.Button` with a row, on all five screens, keeping the order of the entries and what each one does. Verify by capturing Main, Host Game and Options and confirming Start, Back, Connect, Join as spectator and Cancel are rows carrying their position numbers.
- [x] 4.3 Move the widget labels off `rlg.Label` onto the menu's own text, so the label column matches the rows. Verify in the Host Game and Options captures that the labels sit level with their rows in the menu's reds.

## 5. Input, and the widgets that remain

- [x] 5.1 Decide the highlight once a frame, before the frame is composed: W/S and the arrows wrapping, Enter or Space taking the row, the mouse taking the highlight when it moves over one, and only on the main screen. Verify `go build ./...` and `go test ./` pass, and that the Options capture shows the row under the resting mouse lit on a screen that has no keyboard highlight.
- [x] 5.2 Route Escape through the same exit as the Back row on every other screen, clearing any field left in edit mode. Verify by reading the exit against each screen's Back row: error cleared for Host and Join, settings saved for Options, the attempt dropped for Connecting.
- [x] 5.3 Paint the remaining raygui widgets in the menu's reds through `DEFAULT`, and put the editor's dark style back in `enterLevel` as `leaveMatch` already did. Verify by capturing the Host Game and Options screens with the fields, spinners, slider, toggles and selector all in the menu's reds, and by capturing a single-player match to confirm the level still draws in its own palette.
- [x] 5.4 Run `go test ./` and `go test -tags release ./` and confirm both pass, including the existing `TestMenuLayout` unchanged and the new panel, rain and extended red-set tests. Confirm `gofmt -l` is clean.

## 6. What a capture cannot answer

- [ ] 6.1 Play the menu (`go run ./cmd`) and work the main screen with the keyboard alone: W/S and the arrows wrapping at both ends, Enter and Space taking the row, and the highlight following the mouse the moment it moves over another row without the two ever both being lit.
- [ ] 6.2 Press Escape on Host Game with an error shown, on Join Game mid-edit of the address, on Options after moving the music volume, and on Connecting; confirm each leaves exactly as its Back row does and that no field comes back still being edited.
- [ ] 6.3 Start a match, open the level editor and confirm its widgets are in the editor's dark style, not the menu's reds.
- [x] 6.4 Tune the rain's weight and the sweeping band's speed and alpha by eye in play — they read differently in motion than in a capture — and record any values that change in design.md, closing its open question. The rain was judged too faint in the release build and was raised (streaks, body alpha, length, and a brighter tip of its own); the new weight was checked in a capture of the development build. The band was left as it was. Verified by the values recorded in design.md and by `go test ./` passing with the rain's tip named in the scene's red-set test.
- [ ] 6.5 Build the release binary with `.\build.ps1` and confirm it packs, then run it and confirm the menu is the same as the development build's.

## Before archiving

- [ ] Merge the `main-menu` delta by hand, not by replacing blocks: `add-menu-music-and-options` (archived 2026-09-23) rewrote "Pixel-art menu scene" and "Menu widgets fit inside the menu panel" after this delta was written - the virtual screen instead of the window, and the Options screen with its fullscreen, window size and music volume rows. Keep those edits alongside this change's; verify with `openspec validate --specs --strict`
