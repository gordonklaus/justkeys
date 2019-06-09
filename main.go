package main

import (
	"github.com/gordonklaus/ui"
)

func main() { ui.Run(Main) }

func Main() {
	startAudio()
	kbd := NewKeyboard()
	ui.NewWindow(ui.Size{300, 200}, kbd)
}
