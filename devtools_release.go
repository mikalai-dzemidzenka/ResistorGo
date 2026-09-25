//go:build release

package resistor

// DevTools is false in a release build. Every developer-only path is guarded
// by it, so the compiler drops the debug overlay and the
// network-simulation keys from the shipped binary. The level editor ships.
const DevTools = false
