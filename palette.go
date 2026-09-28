package resistor

import (
	"bytes"
	"image/color"
	"image/png"
	"math"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Palette is what the level quantizes to: every Endesga 64 colour a block
// texture's or character sprite's pixels come nearest to, read once at load
// time, plus the colours the HUD and effects must keep exactly.
var Palette = withMidpoints(levelPalette(append(blockTextures(), spriteSheets()...)), 64)

// withMidpoints adds the colour halfway between every two entries of pal
// closer than maxDist, so a tone between two ramps (a warm dark grey between
// Endesga's blue-greys and browns) finds its own entry instead of the nearest
// blue. The entries of pal come first and are unchanged.
func withMidpoints(pal []rl.Color, maxDist int) []rl.Color {
	out := slices.Clone(pal)
	for i, a := range pal {
		for _, b := range pal[i+1:] {
			dr, dg, db := int(a.R)-int(b.R), int(a.G)-int(b.G), int(a.B)-int(b.B)
			if dr*dr+dg*dg+db*db >= maxDist*maxDist {
				continue
			}
			m := rl.Color{uint8((int(a.R) + int(b.R)) / 2), uint8((int(a.G) + int(b.G)) / 2), uint8((int(a.B) + int(b.B)) / 2), 255}
			if !slices.Contains(out, m) {
				out = append(out, m)
			}
		}
	}
	return out
}

// spriteSheets is every character animation sheet, in every skin.
func spriteSheets() []string {
	var names []string
	for _, p := range animTable {
		for _, s := range skins {
			names = append(names, skinSheet(p, s))
		}
	}
	return names
}

// levelPalette builds the level palette from the textures at paths. A
// texture that cannot be read or decoded adds nothing.
func levelPalette(paths []string) []rl.Color {
	var pal []rl.Color
	for _, c := range slices.Concat(bloodReds[:], []rl.Color{attackOutlineColor, redTeamColor, blueTeamColor,
		structureBarTrack, structureBarFill, stunPipColor, energyBarFill,
		// black, or an unlit pixel finds Endesga's dark blue-purples before its
		// darkest grey (19,19,19), which is farther from black than they are
		rl.Black,
		// the ends of the grey ramp, so lighting can fall to near black and
		// flare to white
		EndesgaPalette[1], EndesgaPalette[2], EndesgaPalette[8],
		// dark purples, for shadows with some colour in them
		EndesgaPalette[16], EndesgaPalette[53], EndesgaPalette[54],
		// warm light: ambers and yellows, up to a cream highlight
		EndesgaPalette[24], EndesgaPalette[30], EndesgaPalette[31], EndesgaPalette[32], EndesgaPalette[23],
		// neutral greys halfway along Endesga's grey ramp, so dimly lit walls
		// find a grey before a blue. Not Endesga: it has no more neutral greys.
		{76, 76, 76, 255}, {113, 113, 113, 255}, {156, 156, 156, 255}, {200, 200, 200, 255}}) {
		if !slices.Contains(pal, c) {
			pal = append(pal, c)
		}
	}
	used := make([]bool, len(EndesgaPalette))
	for _, p := range paths {
		data, err := textureBytes(p)
		if err != nil {
			continue
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			continue
		}
		b := img.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
				if c.A == 0 {
					continue
				}
				used[nearestEndesga(c.R, c.G, c.B)] = true
			}
		}
	}
	for i, c := range EndesgaPalette {
		if used[i] && !slices.Contains(pal, c) {
			pal = append(pal, c)
		}
	}
	return pal
}

// nearestEndesga is the index of the Endesga 64 colour closest to r,g,b, by the
// same RGB distance the palette shader uses.
func nearestEndesga(r, g, b uint8) int {
	best, bestD := 0, math.MaxInt
	for i, p := range EndesgaPalette {
		dr, dg, db := int(r)-int(p.R), int(g)-int(p.G), int(b)-int(p.B)
		if d := dr*dr + dg*dg + db*db; d < bestD {
			best, bestD = i, d
		}
	}
	return best
}

// EndesgaPalette is the Endesga 64 palette (https://lospec.com/palette-list/endesga-64),
// in its published order. The main menu is quantized to all 64 colours; the
// level to the ones its blocks use (Palette above).
var EndesgaPalette = []rl.Color{
	{255, 0, 64, 255}, {19, 19, 19, 255}, {27, 27, 27, 255}, {39, 39, 39, 255},
	{61, 61, 61, 255}, {93, 93, 93, 255}, {133, 133, 133, 255}, {180, 180, 180, 255},
	{255, 255, 255, 255}, {199, 207, 221, 255}, {146, 161, 185, 255}, {101, 115, 146, 255},
	{66, 76, 110, 255}, {42, 47, 78, 255}, {26, 25, 50, 255}, {14, 7, 27, 255},
	{28, 18, 28, 255}, {57, 31, 33, 255}, {93, 44, 40, 255}, {138, 72, 54, 255},
	{191, 111, 74, 255}, {230, 156, 105, 255}, {246, 202, 159, 255}, {249, 230, 207, 255},
	{237, 171, 80, 255}, {224, 116, 56, 255}, {198, 69, 36, 255}, {142, 37, 29, 255},
	{255, 80, 0, 255}, {237, 118, 20, 255}, {255, 162, 20, 255}, {255, 200, 37, 255},
	{255, 235, 87, 255}, {211, 252, 126, 255}, {153, 230, 95, 255}, {90, 197, 79, 255},
	{51, 152, 75, 255}, {30, 111, 80, 255}, {19, 76, 76, 255}, {12, 46, 68, 255},
	{0, 57, 109, 255}, {0, 105, 170, 255}, {0, 152, 220, 255}, {0, 205, 249, 255},
	{12, 241, 255, 255}, {148, 253, 255, 255}, {253, 210, 237, 255}, {243, 137, 245, 255},
	{219, 63, 253, 255}, {122, 9, 250, 255}, {48, 3, 217, 255}, {12, 2, 147, 255},
	{3, 25, 63, 255}, {59, 20, 67, 255}, {98, 36, 97, 255}, {147, 56, 143, 255},
	{202, 82, 201, 255}, {200, 80, 134, 255}, {246, 129, 135, 255}, {245, 85, 93, 255},
	{234, 50, 60, 255}, {196, 36, 48, 255}, {137, 30, 43, 255}, {87, 28, 39, 255},
}
