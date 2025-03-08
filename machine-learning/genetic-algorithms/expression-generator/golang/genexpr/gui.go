package genexpr

import (
	"fmt"
	"github.com/AllenDang/imgui-go"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"image/color"
	"math/big"
	"sort"
	"time"

	g "github.com/AllenDang/giu"
)

var (
	headerFont         *g.FontInfo
	chromosomeFontSize float32 = 10
	chromosomeFont     *g.FontInfo
)

type colorThreshold struct {
	threshold float64
	color     color.Color
}

var valueColorThresholds = []colorThreshold{
	{0.0, color.RGBA{R: 255, A: 255}},
	{0.1, color.RGBA{R: 255, G: 128, A: 255}},
	{0.5, color.RGBA{R: 255, G: 255, A: 255}},
	{0.96, color.RGBA{G: 179, A: 255}},
	{0.97, color.RGBA{G: 255, A: 255}},
}
var minValueColor = valueColorThresholds[0].color
var maxValueColor = valueColorThresholds[len(valueColorThresholds)-1].color

var numPrinter = message.NewPrinter(language.English)

type GuiParams struct {
	// Max UI updates to perform per second
	FramesPerSecond uint

	// Artificial delay to add between Simulation steps
	StepDelay time.Duration
}

func DefaultGuiParams() *GuiParams {
	return &GuiParams{
		FramesPerSecond: 60,
		StepDelay:       time.Nanosecond,
	}
}

func initFont() {
	headerFont = g.AddFontFromBytes("gomono.ttf", gomono.TTF, 20)
	chromosomeFont = g.AddFontFromBytes("gomono.ttf", gomono.TTF, chromosomeFontSize)
}

func renderChromosome(popMember *PopulationMember, maxResultLen int) g.Widget {
	c := popMember.Chromosome()
	decoded := c.Decode()
	result, err := decoded.Evaluate()

	resultStr := "ERROR"
	resultColor := color.Color(color.RGBA{R: 255, A: 255})
	if err == nil {
		resultStr = numPrinter.Sprintf("% .2f", result)
		fitnessFloat, _ := popMember.Fitness().Float64()

		colorIndex := sort.Search(len(valueColorThresholds), func(i int) bool {
			return valueColorThresholds[i].threshold >= fitnessFloat
		})
		if colorIndex >= len(valueColorThresholds) {
			resultColor = maxValueColor
		} else if colorIndex == 0 {
			resultColor = minValueColor
		} else {
			resultColor = valueColorThresholds[colorIndex].color

			fromIndex := colorIndex - 1
			fromThresholdColor := valueColorThresholds[fromIndex]
			toThresholdColor := valueColorThresholds[colorIndex]

			distance := (fitnessFloat - fromThresholdColor.threshold) / (toThresholdColor.threshold - fromThresholdColor.threshold)
			distance *= 2

			var red, green, blue uint8
			comps := []struct {
				from   uint8
				to     uint8
				result *uint8
			}{
				{fromThresholdColor.color.(color.RGBA).R, toThresholdColor.color.(color.RGBA).R, &red},
				{fromThresholdColor.color.(color.RGBA).G, toThresholdColor.color.(color.RGBA).G, &green},
				{fromThresholdColor.color.(color.RGBA).B, toThresholdColor.color.(color.RGBA).B, &blue},
			}
			for _, comp := range comps {
				diff := comp.to - comp.from
				*comp.result = comp.from + uint8(float64(diff)*distance)
			}

			resultColor = color.RGBA{R: red, G: green, B: blue, A: 255}
		}
	}

	return g.Style().SetFont(chromosomeFont).To(g.Custom(func() {
		geneLabels := make([]g.Widget, len(c.Genes()))

		geneFmt := fmt.Sprintf("%%0%db", GeneBits)
		genes := c.Genes()
		for i, validity := range decoded.Validity {
			gene := genes[i]
			var geneColor color.Color

			switch validity {
			case rune(Valid):
				geneColor = color.Color(color.RGBA{G: 255, A: 255})
			case rune(Invalid):
				geneColor = color.Color(color.RGBA{R: 255, G: 255, A: 255})
			case rune(Unknown):
				geneColor = color.Color(color.RGBA{R: 255, A: 255})
			}

			geneLabels[i] = g.Style().SetColor(g.StyleColorText, geneColor).To(
				g.Labelf(geneFmt, gene),
			)
		}

		widgets := append(
			geneLabels,
			g.Label("≈"),
			g.Style().SetColor(g.StyleColorText, resultColor).To(
				g.Label(numPrinter.Sprintf("%*s", maxResultLen, resultStr)),
			),
			g.Label("≈"),
			g.Style().SetColor(g.StyleColorText, resultColor).To(
				g.Label(decoded.Expression),
			),
		)
		g.Row(widgets...).Build()
	}))
}

func GuiMain(sim *Simulation, getTarget func() *big.Int, guiParams *GuiParams) {
	target := sim.Target()
	fmt.Printf("Target: %f\n", target)

	guiTicker := time.NewTicker(time.Second / time.Duration(guiParams.FramesPerSecond))
	stepTicker := time.NewTicker(guiParams.StepDelay)

	didSolve := false
	isPopulationDirty := true
	stickyMaxResultLen := 0
	framesSinceStickyMaxResultLenChanged := 0
	needsWindowResize := true

	var wnd *g.MasterWindow

	loop := func() {
		var targetInt big.Int
		target.Int(&targetInt)

		if needsWindowResize {
			g.PushFont(chromosomeFont)
			// imgui appears to add padding on the left/right of chars, so we can determine the
			// *true* width of a char by measuring the size difference between a 1-char and 2-char string
			oneCharWidth, _ := g.CalcTextSize(" ")
			twoCharWidth, _ := g.CalcTextSize("  ")
			charWidth := twoCharWidth - oneCharWidth
			g.PopFont()

			winWidth, winHeight := calculateWindowSize(charWidth, sim)
			wnd.SetSize(winWidth, winHeight)

			needsWindowResize = false
		}

		g.SingleWindow().Layout(
			g.Row(
				g.Style().SetFont(headerFont).To(
					g.Row(
						g.Custom(func() {
							equalityOp := "≈"
							if didSolve {
								equalityOp = "="
							}

							var topChromosome *Chromosome
							if len(sim.Solutions()) > 0 {
								topChromosome = sim.Solutions()[0]
							} else {
								sortedPopulation := sim.GetSortedPopulation()
								topChromosome = sortedPopulation[0].Chromosome()
							}

							g.Align(g.AlignLeft).To(
								g.Label(numPrinter.Sprintf("%s %s %s", targetInt.String(), equalityOp, topChromosome.Decode().Expression)),
							).Build()
						}),
						g.AlignManually(g.AlignRight, g.Label(numPrinter.Sprintf("Iteration: %d", sim.Iteration())), 200, false),
					),
				),
			),
			g.Row(
				g.Column(g.Custom(func() {
					maxResultLen := 0
					for _, member := range sim.Population() {
						result, err := member.Chromosome().Decode().Evaluate()
						if err != nil {
							continue
						}

						resultLen := len(numPrinter.Sprintf("%+.2f", result))
						if resultLen > maxResultLen {
							maxResultLen = resultLen
						}
					}

					if didSolve || maxResultLen > stickyMaxResultLen {
						stickyMaxResultLen = maxResultLen
						framesSinceStickyMaxResultLenChanged = 0
					} else if maxResultLen < stickyMaxResultLen && framesSinceStickyMaxResultLenChanged > 30 {
						stickyMaxResultLen--
						framesSinceStickyMaxResultLenChanged = 0
					} else {
						framesSinceStickyMaxResultLenChanged++
					}

					for _, member := range sim.Population() {
						renderChromosome(member, stickyMaxResultLen).Build()
					}
				})),
			),
		)
	}

	go func() {
		for {
			select {
			case <-guiTicker.C:
				if isPopulationDirty {
					isPopulationDirty = false
					g.Update()
				}
			case <-stepTicker.C:
				if didSolve {
					continue
				}

				didSolve = sim.Step()
				isPopulationDirty = true

				if didSolve {
					fmt.Println("SOLVED!")
					for _, c := range sim.Solutions() {
						fmt.Println(c.Decode().Expression)
					}
				}
			}
		}
	}()

	restartSim := func(reuseTarget bool) {
		var targetInt big.Int
		if reuseTarget {
			target.Int(&targetInt)
		} else {
			targetInt.Set(getTarget())
			target.SetInt(&targetInt)
		}

		sim = NewSimulation(sim.Params())
		sim.Init(&targetInt)

		didSolve = false
		isPopulationDirty = true
	}

	initFont()

	winWidth, winHeight := calculateWindowSize(chromosomeFontSize, sim)
	wnd = g.NewMasterWindow("Genetic Expression Calculator", winWidth, winHeight, 0)

	// Avoid "Too many vertices in ImDrawList using 16-bit indices" assertion
	g.Context.IO().SetBackendFlags(imgui.BackendFlagsRendererHasVtxOffset)

	wnd.RegisterKeyboardShortcuts(
		g.WindowShortcut{Key: g.KeyEnter, Modifier: 0, Callback: func() {
			restartSim(false)
		}},
		g.WindowShortcut{Key: g.KeyEnter, Modifier: g.ModControl, Callback: func() {
			restartSim(true)
		}},
	)
	wnd.Run(loop)
}

func calculateWindowSize(charWidth float32, sim *Simulation) (int, int) {
	chromosomeSize := float32(sim.Params().ChromosomeSize)

	geneWidth := charWidth*float32(GeneBits) + 4
	chromosomeWidth := geneWidth*chromosomeSize + charWidth*(chromosomeSize-1)
	maxExprWidth := charWidth * chromosomeSize

	return int(chromosomeWidth + maxExprWidth*1.4 + 100), sim.Population().Len()*14 + 100
}
