package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/mazznoer/colorgrad"
	"image/color"
	"log"
	"math"
	"sync"
)

const (
	defaultScreenWidth  = 1024
	defaultScreenHeight = 768

	maxIterations = 1_500

	zoomPercent = 0.1
	zoomScale = 1 - zoomPercent

	mandelbrotMinX, mandelbrotMaxX = -2.5, 1.0
	mandelbrotMinY, mandelbrotMaxY = -1.0, 1.0
)

type Game struct {
	win *pixelgl.Window

	canvas *pixel.PictureData
	canvasBounds pixel.Rect
	scale pixel.Matrix
	grad colorgrad.Gradient

	numRenderWorkers int
	cancelRender chan struct{}
	isRendering bool
	scheduleRenderLock sync.Mutex

	needsRedraw bool
}

func newMatrix(a, b, tx, c, d, ty float64) pixel.Matrix {
	return pixel.Matrix{
		0: a, 2: b, 4: tx,
		1: c, 3: d, 5: ty,
	}
}

func NewGame(win *pixelgl.Window) *Game {
	win.MakePicture(pixel.MakePictureData(win.Bounds()))

	grad, _ := colorgrad.NewGradient().
		HtmlColors("#000635", "#13399a", "#1360d4", "#ffffff", "#ff7f0e", "#653608", "#000000").
		Mode(colorgrad.BlendHcl).
		Domain(0, 2, 10, 30, 100, 200, maxIterations).
		Build()

	g := &Game{
		win: win,

		needsRedraw:      false,
		cancelRender:     make(chan struct{}),
		numRenderWorkers: 12,

		scale: newMatrix(
			mandelbrotMaxX-mandelbrotMinX, 0, mandelbrotMinX,
			0, mandelbrotMaxY-mandelbrotMinY, mandelbrotMinY,
		),
		grad: grad,
	}
	g.Resize(win.Bounds())
	return g
}

func (g *Game) Resize(bounds pixel.Rect) {
	log.Println("Resized to: ", bounds.String())
	g.canvas = pixel.MakePictureData(bounds)
	g.canvasBounds = bounds
}

func (g *Game) scheduleRender() {
	scale := g.scale

	g.scheduleRenderLock.Lock()
	defer g.scheduleRenderLock.Unlock()

	if g.isRendering {
		g.cancelRender <- struct{}{}
		<-g.cancelRender
	}

	go g.renderMandelbrot(scale, g.canvas, g.canvasBounds)
}

func (g *Game) Update() {
	if g.win.Bounds() != g.canvasBounds {
		g.Resize(g.win.Bounds())
		g.scheduleRender()
	}

	scroll := g.win.MouseScroll()

	if scroll.Y > 0 {
		scaleFactor := math.Pow(zoomScale, scroll.Y)
		g.applyZoom(scaleFactor, scaleFactor)
	} else if scroll.Y < 0 {
		scaleFactor := math.Pow(1 + zoomPercent, -scroll.Y)
		g.applyZoom(scaleFactor, scaleFactor)
	}
}

func (g *Game) applyZoom(xZoomScale, yZoomScale float64) {
	mousePos := g.win.MousePosition()
	mouseScale := mousePos.ScaledXY(pixel.V(1.0 / g.canvasBounds.Max.X, 1.0 / g.canvasBounds.Max.Y))
	mouseMandel := g.scale.Project(mouseScale)

	g.scale = g.scale.ScaledXY(mouseMandel, pixel.V(xZoomScale, yZoomScale))
	g.scheduleRender()
}

// renderMandelbrot draws the brush on the given canvas image at the position (x, y).
func (g *Game) renderMandelbrot(scale pixel.Matrix, canvas *pixel.PictureData, bounds pixel.Rect) {
	g.isRendering = true
	defer func() { g.isRendering = false }()

	iScreenWidth := int(bounds.Max.X)
	iScreenHeight := int(bounds.Max.Y)

	wg := sync.WaitGroup{}

	lineWorker := func(yChan <-chan int) {
		defer wg.Done()

		for screenY := range yChan {
			yScale := float64(screenY) / bounds.Max.Y

			for screenX := 0; screenX < iScreenWidth; screenX++ {
				xScale := float64(screenX) / bounds.Max.X
				coord := scale.Project(pixel.V(xScale, yScale))

				numIterations := 0

				offset := complex(coord.X, coord.Y)
				n := offset
				for ; real(n)*real(n) + imag(n)*imag(n) <= 2*2 && numIterations < maxIterations; numIterations++ {
					n = n*n + offset
				}

				iterColor := g.grad.At(float64(numIterations))
				red, green, blue := iterColor.RGB255()

				canvas.Pix[screenY * canvas.Stride + screenX] = color.RGBA{R: red, G: green, B: blue, A: 255}
			}
		}
	}

	yChan := make(chan int)
	for i := 0; i < g.numRenderWorkers; i++ {
		wg.Add(1)
		go lineWorker(yChan)
	}

	didCancelRender := false

	lineLoop:
	for screenY := 0; screenY < iScreenHeight; screenY++ {
		select {
		case yChan <- screenY:
		case <-g.cancelRender:
			didCancelRender = true
			break lineLoop
		}
	}
	close(yChan)

	wg.Wait()

	if didCancelRender {
		log.Println("Canceled render call")
		// Inform our canceler that we're done
		g.cancelRender <- struct{}{}
	} else {
		g.needsRedraw = true
	}
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title: "Mandelbrot Set Viz",
		Bounds: pixel.R(0, 0, defaultScreenWidth, defaultScreenHeight),
		VSync: true,
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	g := NewGame(win)
	g.scheduleRender()

	for !win.Closed() {
		g.Update()

		if g.needsRedraw {
			win.Clear(color.Black)
			sprite := pixel.NewSprite(g.canvas, g.canvas.Bounds())
			sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
			g.needsRedraw = false
		}

		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
