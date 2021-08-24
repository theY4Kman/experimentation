package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image/color"
	"log"
	"math"
	"sync"
)

const (
	screenWidth  = 1024
	screenHeight = 768

	maxIterations = 1_000

	zoomPercent = 0.1
	zoomScale = 1 - zoomPercent

	mandelbrotMinX, mandelbrotMaxX = -2.5, 1.0
	mandelbrotMinY, mandelbrotMaxY = -1.0, 1.0
)

type Game struct {
	count int

	scale ebiten.GeoM

	needsRedraw bool
	canvasImage *ebiten.Image
}

func NewGame() *Game {
	g := &Game{
		canvasImage: ebiten.NewImage(screenWidth, screenHeight),
		needsRedraw: true,

		// NOTE: as screen X increases, so does our mandelbrot X coord,
		//		 but as screen Y increases, our mandelbrot Y *decreases*
		scale: newGeoM(
			mandelbrotMaxX - mandelbrotMinX, 0, mandelbrotMinX,
			0, mandelbrotMinY - mandelbrotMaxY, mandelbrotMaxY,
		),
	}
	return g
}

// newGeoM initializes a transformation matrix in the form:
//  [ a  c
//    b  d
//    tx ty ]
func newGeoM(a, b, tx, c, d, ty float64) ebiten.GeoM {
	outp := ebiten.GeoM{}
	outp.SetElement(0, 0, a)
	outp.SetElement(0, 1, b)
	outp.SetElement(0, 2, tx)
	outp.SetElement(1, 0, c)
	outp.SetElement(1, 1, d)
	outp.SetElement(1, 2, ty)
	return outp
}

func (g *Game) Update() error {
	_, scrollY := ebiten.Wheel()

	if scrollY > 0 {
		scaleFactor := math.Pow(zoomScale, scrollY)
		g.applyZoom(scaleFactor, scaleFactor)
	} else if scrollY < 0 {
		scaleFactor := math.Pow(1 + zoomPercent, -scrollY)
		g.applyZoom(scaleFactor, scaleFactor)
	}

	if g.needsRedraw {
		g.canvasImage.Fill(color.Black)
		g.renderMandelbrot(g.canvasImage)

		// Draw current mandelbrot x,y bounds
		x0, y0 := g.scale.Apply(0, 0)
		x1, y1 := g.scale.Apply(1, 1)
		ebitenutil.DebugPrint(g.canvasImage, fmt.Sprintf("(%f, %f) to (%f, %f)", x0, y0, x1, y1))

		g.needsRedraw = false
		g.count++
	}

	return nil
}

func (g *Game) applyZoom(xZoomScale, yZoomScale float64) {
	var zoomMatrix ebiten.GeoM

	mouseX, mouseY := ebiten.CursorPosition()
	mouseXScale := float64(mouseX) / float64(screenWidth)
	mouseYScale := float64(mouseY) / float64(screenHeight)

	mouseMX, mouseMY := g.scale.Apply(mouseXScale, mouseYScale)

	zoomMatrix = newGeoM(1, 0, -mouseMX, 0, 1, -mouseMY)
	zoomMatrix.Scale(xZoomScale, yZoomScale)
	zoomMatrix.Translate(mouseMX, mouseMY)

	g.scale.Concat(zoomMatrix)
	g.needsRedraw = true
}

// renderMandelbrot draws the brush on the given canvas image at the position (x, y).
func (g *Game) renderMandelbrot(canvas *ebiten.Image) {
	pixels := make([]byte, 4 * screenWidth * screenHeight)

	colorBase := math.Log(float64(maxIterations))
	colorStep := float64(20)
	fMaxRawColor := float64(^uint32(0) >> 8)

	wg := sync.WaitGroup{}

	for screenY := 0; screenY < screenHeight; screenY++ {
		wg.Add(1)

		go func(screenY int) {
			yScale := float64(screenY) / float64(screenHeight)

			for screenX := 0; screenX < screenWidth; screenX++ {
				xScale := float64(screenX) / float64(screenWidth)

				x0, y0 := g.scale.Apply(xScale, yScale)

				i := 0
				x, y := 0.0, 0.0
				for ; x*x + y*y < 2*2 && i < maxIterations; i++ {
					x, y = x*x - y*y + x0, 2*x*y + y0
				}

				colorScale := math.Log(float64(i)) / colorBase
				rawColor := uint32(math.Round(fMaxRawColor * colorScale / colorStep) * colorStep)

				pixIndex := 4 * (screenY * screenWidth + screenX)
				pixels[pixIndex] = byte(rawColor & 0xff)
				pixels[pixIndex+1] = byte((rawColor >> 8) & 0xff)
				pixels[pixIndex+2] = byte((rawColor >> 16) & 0xff)
				pixels[pixIndex+3] = 255
			}

			wg.Done()
		}(screenY)
	}

	wg.Wait()
	canvas.ReplacePixels(pixels)
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.canvasImage, nil)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Mandelbrot Set")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
