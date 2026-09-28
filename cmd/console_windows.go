package main

import (
	"syscall"
	"unsafe"
)

// hideConsole lets go of the console when the game is its only user, as it is
// when launched from Explorer: Windows then closes the window. Started from a
// terminal the console is shared with the shell, and is kept for the log.
// ponytail: the window still flashes up at launch; linking with -H windowsgui
// avoids that, but -dedicated would then need its own console attached.
func hideConsole() {
	k := syscall.NewLazyDLL("kernel32.dll")
	var pids [2]uint32
	n, _, _ := k.NewProc("GetConsoleProcessList").Call(uintptr(unsafe.Pointer(&pids[0])), 2)
	if n == 1 {
		k.NewProc("FreeConsole").Call()
	}
}
