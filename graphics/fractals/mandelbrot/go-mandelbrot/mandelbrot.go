package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/mazznoer/colorgrad"
	"image/color"
	"log"
	"math"
	"math/cmplx"
	"sync"
)

const (
	defaultScreenWidth  = 1024
	defaultScreenHeight = 768

	maxIterations = 1_500

	zoomPercent = 0.1
	zoomScale   = 1 - zoomPercent

	mandelbrotMinX, mandelbrotMaxX = -2.5, 1.0
	mandelbrotMinY, mandelbrotMaxY = -1.0, 1.0
)

type MandelbrotFunc func(complex128) int

func (f MandelbrotFunc) calculate(offset complex128) int {
	return f(offset)
}

func mandelbrotNaiveComplexCond(offset complex128) int {
	numIterations := 0
	for n := offset; cmplx.Abs(n) <= 2 && numIterations < maxIterations; numIterations++ {
		n = n*n + offset
	}
	return numIterations
}

func mandelbrotNaiveFloatCond(offset complex128) int {
	numIterations := 0
	for n := offset; real(n)*real(n)+imag(n)*imag(n) <= 2*2 && numIterations < maxIterations; numIterations++ {
		n = n*n + offset
	}
	return numIterations
}

func mandelbrotOptimizedEscape1(offset complex128) int {
	numIterations := 0
	r2, i2 := 0.0, 0.0
	w := 0.0
	r, i := real(offset), imag(offset)
	for ; r2+i2 <= 2*2 && numIterations < maxIterations; numIterations++ {
		r = r2 - i2 + real(offset)
		i = w - r2 - i2 + imag(offset)
		r2 = r * r
		i2 = i * i
		w = (r + i) * (r + i)
	}
	return numIterations
}

func mandelbrotOptimizedEscape2(offset complex128) int {
	numIterations := 0
	r2, i2 := 0.0, 0.0
	w := 0.0
	r, i := real(offset), imag(offset)
	for ; r2+i2 <= 2*2 && numIterations < maxIterations; numIterations++ {
		r = r2 - i2 + real(offset)
		i = w - r2 - i2 + imag(offset)
		r2 = r * r
		i2 = i * i
		w = r2 + 2*r*i + i2
	}
	return numIterations
}

func mandelbrotOptimizedEscape3(offset complex128) int {
	numIterations := 0
	r2, i2 := 0.0, 0.0
	r, i := real(offset), imag(offset)
	for ; r2+i2 <= 2*2 && numIterations < maxIterations; numIterations++ {
		i = 2*r*i + imag(offset)
		r = r2 - i2 + real(offset)
		r2 = r * r
		i2 = i * i
	}
	return numIterations
}

//func mandelbrot(offset complex128) int {
//	startedAt := time.Now()
//	numIterations := mandelbrotOptimizedEscape3(offset)
//	log.Printf("complex(%f,%f),%d", real(offset), imag(offset), time.Since(startedAt).Milliseconds())
//	return numIterations
//}

var mandelbrot MandelbrotFunc = mandelbrotOptimizedEscape3

type MandelbrotViz struct {
	win *pixelgl.Window

	canvas       *pixel.PictureData
	canvasBounds pixel.Rect
	scale        pixel.Matrix
	grad         colorgrad.Gradient

	numRenderWorkers   int
	cancelRender       chan struct{}
	didCancelRender    chan struct{}
	isRendering        bool
	scheduleRenderLock sync.Mutex

	needsRedraw bool
}

func newMatrix(a, b, tx, c, d, ty float64) pixel.Matrix {
	return pixel.Matrix{
		0: a, 2: b, 4: tx,
		1: c, 3: d, 5: ty,
	}
}

func NewMandelbrotViz(win *pixelgl.Window) *MandelbrotViz {
	win.MakePicture(pixel.MakePictureData(win.Bounds()))

	grad, _ := colorgrad.NewGradient().
		HtmlColors("#000635", "#13399a", "#1360d4", "#ffffff", "#ff7f0e", "#653608", "#000000").
		Mode(colorgrad.BlendHcl).
		Domain(0, 2, 10, 30, 100, 200, maxIterations).
		Build()

	g := &MandelbrotViz{
		win: win,

		needsRedraw:      false,
		cancelRender:     make(chan struct{}),
		didCancelRender:  make(chan struct{}),
		numRenderWorkers: 400,

		scale: newMatrix(
			mandelbrotMaxX-mandelbrotMinX, 0, mandelbrotMinX,
			0, mandelbrotMaxY-mandelbrotMinY, mandelbrotMinY,
		),
		grad: grad,
	}
	g.Resize(win.Bounds())
	return g
}

func (g *MandelbrotViz) Resize(bounds pixel.Rect) {
	log.Println("Resized to: ", bounds.String())
	g.canvas = pixel.MakePictureData(bounds)
	g.canvasBounds = bounds
}

func (g *MandelbrotViz) scheduleRender() {
	scale := g.scale

	g.scheduleRenderLock.Lock()
	defer g.scheduleRenderLock.Unlock()

	if g.isRendering {
		g.cancelRender <- struct{}{}
		<-g.didCancelRender
	}

	go g.renderMandelbrot(scale, g.canvas, g.canvasBounds)
}

func (g *MandelbrotViz) Update() {
	if g.win.Bounds() != g.canvasBounds {
		g.Resize(g.win.Bounds())
		g.scheduleRender()
	}

	scroll := g.win.MouseScroll()

	if scroll.Y > 0 { // scroll up = zoom in
		scaleFactor := math.Pow(zoomScale, scroll.Y)
		g.applyZoom(scaleFactor, scaleFactor)
	} else if scroll.Y < 0 { // scroll down = zoom out
		scaleFactor := math.Pow(1+zoomPercent, -scroll.Y)
		g.applyZoom(scaleFactor, scaleFactor)
	}

	mousePos := g.win.MousePosition()
	yScale := mousePos.Y / g.canvasBounds.Max.Y
	xScale := mousePos.X / g.canvasBounds.Max.X
	coord := g.scale.Project(pixel.V(xScale, yScale))
	g.win.SetTitle(fmt.Sprintf("Mandelbrot Set Viz - %s", coord))

	if g.win.JustPressed(pixelgl.MouseButtonLeft) {
		log.Printf("complex(%f,%f)", coord.X, coord.Y)
	}
}

func (g *MandelbrotViz) applyZoom(xZoomScale, yZoomScale float64) {
	g.scale = g.calculateZoom(g.scale, g.win.MousePosition(), g.canvasBounds.Max, pixel.V(xZoomScale, yZoomScale))
	g.scheduleRender()
}

func (g *MandelbrotViz) calculateZoom(scale pixel.Matrix, mousePos, screenMax, zoomScale pixel.Vec) pixel.Matrix {
	mouseScale := mousePos.ScaledXY(pixel.V(1.0/screenMax.X, 1.0/screenMax.Y))
	mouseMandel := scale.Project(mouseScale)
	return scale.ScaledXY(mouseMandel, zoomScale)
}

// renderMandelbrot draws the brush on the given canvas image at the position (x, y).
func (g *MandelbrotViz) renderMandelbrot(scale pixel.Matrix, canvas *pixel.PictureData, bounds pixel.Rect) {
	g.isRendering = true
	defer func() { g.isRendering = false }()

	iScreenWidth := int(bounds.Max.X)
	iScreenHeight := int(bounds.Max.Y)

	wg := sync.WaitGroup{}

	lineWorker := func(yChan <-chan int) {
		defer wg.Done()

		for screenY := range yChan {
			//XXX///////////////////////////////////////////////////////////////////////////////////////////
			//XXX///////////////////////////////////////////////////////////////////////////////////////////
			//startedAt := time.Now()
			//XXX///////////////////////////////////////////////////////////////////////////////////////////
			yScale := float64(screenY) / bounds.Max.Y
			rowOffset := screenY * canvas.Stride

			for screenX := 0; screenX < iScreenWidth; screenX++ {
				xScale := float64(screenX) / bounds.Max.X
				coord := scale.Project(pixel.V(xScale, yScale))

				numIterations := mandelbrot(complex(coord.X, coord.Y))

				iterColor := g.grad.At(float64(numIterations))
				red, green, blue := iterColor.RGB255()

				canvas.Pix[rowOffset+screenX] = color.RGBA{R: red, G: green, B: blue, A: 255}
			}

			//XXX///////////////////////////////////////////////////////////////////////////////////////////
			//XXX///////////////////////////////////////////////////////////////////////////////////////////
			//log.Printf("Rendered line %d in %s", screenY, time.Since(startedAt))
			//XXX///////////////////////////////////////////////////////////////////////////////////////////
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

	select {
	case <-g.cancelRender:
		didCancelRender = true
	default:
	}

	if didCancelRender {
		log.Println("Canceled render call")
		// Inform our canceler that we're done
		g.didCancelRender <- struct{}{}
	} else {
		g.needsRedraw = true
	}
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:     "Mandelbrot Set Viz",
		Bounds:    pixel.R(0, 0, defaultScreenWidth, defaultScreenHeight),
		VSync:     true,
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	g := NewMandelbrotViz(win)
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
