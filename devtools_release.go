//go:build release

package resistor

// DevTools is false in a release build. Every developer-only path is guarded
// by it, so the compiler drops the editor, the debug overlay, the
// network-simulation keys and all level writing from the shipped binary.
const DevTools = false
