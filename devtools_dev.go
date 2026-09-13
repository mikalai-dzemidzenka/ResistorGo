//go:build !release

package resistor

// DevTools is true in a development build: the level editor, the debug
// overlay and the network-simulation keys are compiled in. build.ps1 builds
// with the release tag, which leaves them out (see devtools_release.go).
const DevTools = true
