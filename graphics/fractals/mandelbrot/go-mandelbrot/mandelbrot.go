package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image/color"
	"log"
	"math"
)

const (
	screenWidth  = 1024
	screenHeight = 768
	aspectRatio = float64(screenWidth) / float64(screenHeight)

	maxIterations = 1000

	zoomPercent = 0.25

	xZoomPercent = zoomPercent
	xZoomScale = 1 - xZoomPercent
	yZoomPercent = xZoomPercent * aspectRatio
	yZoomScale = 1 - yZoomPercent

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
	g.canvasImage.Fill(color.Black)
	g.renderMandelbrot(g.canvasImage)
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
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.applyZoom(xZoomScale, yZoomScale, true)
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		g.applyZoom(1 + xZoomPercent, 1 + yZoomPercent, false)
	}

	if g.needsRedraw {
		g.renderMandelbrot(g.canvasImage)
		g.count++
	}

	return nil
}

func (g *Game) applyZoom(xZoomScale, yZoomScale float64, fromMouse bool) {
	var zoomMatrix ebiten.GeoM

	//XXX///////////////////////////////////////////////////////////////////////////////////////////
	log.Println("Scale before zoom: ", g.scale.String())

	if fromMouse {
		mouseX, mouseY := ebiten.CursorPosition()
		mouseXScale := float64(mouseX) / float64(screenWidth)
		mouseYScale := float64(mouseY) / float64(screenHeight)

		zoomMatrix = newGeoM(1, 0, -mouseXScale, 0, 1, -mouseYScale)
		zoomMatrix.Scale(xZoomScale, yZoomScale)
		zoomMatrix.Translate(mouseXScale, mouseYScale)

		//XXX///////////////////////////////////////////////////////////////////////////////////////////
		log.Println("Zooming from mouse: ", zoomMatrix.String())
	} else {
		//zoomMatrix = newGeoM(0.5, 0, 0, 0, 0.5, 0)
		zoomMatrix.Scale(xZoomScale, yZoomScale)

		//XXX///////////////////////////////////////////////////////////////////////////////////////////
		log.Println("Zooming w/o mouse: ", zoomMatrix.String())
	}

	g.scale.Concat(zoomMatrix)

	//XXX///////////////////////////////////////////////////////////////////////////////////////////
	log.Println("Zoom result: ", g.scale.String())
	fmt.Println()

	g.needsRedraw = true
}

// renderMandelbrot draws the brush on the given canvas image at the position (x, y).
func (g *Game) renderMandelbrot(canvas *ebiten.Image) {

	// Make origin (0,0)
	//scale.Translate(-scale.Element(0, 0), -scale.Element(1, 0))

	for screenY := 0; screenY < screenHeight; screenY++ {
		yScale := float64(screenY) / float64(screenHeight)

		for screenX := 0; screenX < screenWidth; screenX++ {
			xScale := float64(screenX) / float64(screenWidth)

			x0, y0 := g.scale.Apply(xScale, yScale)

			i := 0
			x, y := 0.0, 0.0
			for ; x*x + y*y < 2*2 && i < maxIterations; i++ {
				x, y = x*x - y*y + x0, 2*x*y + y0
			}

			colorScale := math.Log(float64(i)) / math.Log(float64(maxIterations))

			clr := color.Gray16{Y: uint16(float64(^uint16(0)) * colorScale)}
			canvas.Set(screenX, screenY, clr)
		}
	}
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
