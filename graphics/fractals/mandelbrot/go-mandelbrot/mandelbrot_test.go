package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"testing"
)

func TestMandelbrot(t *testing.T) {
	main()
}

func TestGame_renderMandelbrot(t *testing.T) {
	g := NewGame()
	canvas := ebiten.NewImage(screenWidth, screenHeight)

	for i := 0; i<10; i++ {
		g.renderMandelbrot(canvas)
	}
}
