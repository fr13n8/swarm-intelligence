package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 720
	ScreenHeight = 576
)

func main() {
	world := NewWorld(ScreenWidth, ScreenHeight)
	world.Init()

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Swarm intelligence!")

	if err := ebiten.RunGame(world); err != nil {
		panic(err)
	}
}
