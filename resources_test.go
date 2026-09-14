package resistor

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// The picker's order must not depend on which source a texture came from, or
// the same entry would move between runs.
func TestMergeTextureListsSortsAndDedupes(t *testing.T) {
	embedded := []string{blocksDir + "/b.png", blocksDir + "/a.png"}
	disk := []string{blocksDir + "/c.png", blocksDir + "/a.png"}
	want := []string{blocksDir + "/a.png", blocksDir + "/b.png", blocksDir + "/c.png"}

	got := mergeTextureLists(embedded, disk)
	if !slices.Equal(got, want) {
		t.Errorf("merge = %v, want %v", got, want)
	}
	if swapped := mergeTextureLists(disk, embedded); !slices.Equal(swapped, want) {
		t.Errorf("the order of the sources changed the result: %v", swapped)
	}
	if none := mergeTextureLists(nil, nil); len(none) != 0 {
		t.Errorf("merge of nothing = %v, want empty", none)
	}
}

// Every block texture in the folder is placeable, not only the three that had
// rectangles written in Go.
func TestBlockTexturesIsEveryBlockTexture(t *testing.T) {
	entries, err := os.ReadDir(blocksDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	var want []string
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".png" {
			want = append(want, blocksDir+"/"+e.Name())
		}
	}
	slices.Sort(want)

	got := blockTextures()
	if !slices.Equal(got, want) {
		t.Errorf("blockTextures = %v, want %v", got, want)
	}
	if len(got) <= len(LoadLevelObjects()) {
		t.Errorf("blockTextures returned %d textures, no more than the %d with built-in rectangles", len(got), len(LoadLevelObjects()))
	}
}

// A texture only on disk can be placed in a development build, but the editor
// has to know it would not survive a build.
func TestInBuildFollowsTheEmbeddedSet(t *testing.T) {
	if !inBuild(blocksDir + "/e1m1_wall1.png") {
		t.Errorf("a texture embedded in the binary is reported as not in the build")
	}
	if inBuild(blocksDir + "/not_a_real_texture.png") {
		t.Errorf("a texture that does not exist is reported as in the build")
	}
	for _, p := range embeddedBlockPaths() {
		if !inBuild(p) {
			t.Errorf("embedded but not in the build: %s", p)
		}
	}
}

// A release build plays with what it was built with: it must not read the
// textures folder at all.
func TestDiskBlockPathsOnlyInDevBuild(t *testing.T) {
	got := diskBlockPaths()
	if !DevTools && got != nil {
		t.Errorf("a release build listed textures on disk: %v", got)
	}
	if DevTools && len(got) == 0 {
		t.Errorf("a development build found no textures in %s", blocksDir)
	}
}

// Garbage must not reach the texture registry: the import path asks this
// before it copies anything.
func TestImageDecodable(t *testing.T) {
	png, err := assetFS.ReadFile(blocksDir + "/e1m1_wall1.png")
	if err != nil {
		t.Fatal(err)
	}
	if !imageDecodable(".png", png) {
		t.Errorf("a real PNG was rejected")
	}
	if imageDecodable(".png", []byte("this is not a png")) {
		t.Errorf("garbage was accepted as an image")
	}
	if imageDecodable(".png", nil) {
		t.Errorf("no bytes at all were accepted as an image")
	}
}
