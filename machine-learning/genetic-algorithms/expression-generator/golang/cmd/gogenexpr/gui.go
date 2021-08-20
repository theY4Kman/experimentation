package main

import (
	"fmt"
	"github.com/AllenDang/imgui-go"
	"github.com/they4kman/experimentation/machine-learning/genetic-algorithms/expression-generator/golang/genexpr"
	"time"

	g "github.com/AllenDang/giu"
)

var sim *genexpr.Simulation

var geneFont imgui.Font
var populationLabels []g.Widget

func initFont() {
	fonts := g.Context.IO().Fonts()

	cfg := imgui.NewFontConfig()
	cfg.SetSize(10)
	geneFont = fonts.AddFontDefaultV(cfg)
}

func loop() {
	refreshPopulation()
	g.SingleWindow("Demo").Layout(populationLabels...)
}

func refreshPopulation() {
	//XXX///////////////////////////////////////////////////////////////////////////////////////////
	fmt.Println("Refreshing population")

	g.PushFont(geneFont)
	for i, member := range sim.Population() {
		populationLabels[i] = g.Label(member.Chromosome().String())
	}
	g.PopFont()
}

func guiMain() {
	params := genexpr.DefaultSimulationParams()
	sim = genexpr.NewSimulation(params)
	sim.InitFromInt(987654322452312411)
	sim.Step()

	populationLabels = make([]g.Widget, params.PopulationSize)
	//refreshPopulationChan = make(chan interface{}, 1)

	go func() {
		guiTick := time.NewTicker(250 * time.Millisecond)
		stepTick := time.NewTicker(10 * time.Millisecond)

		for {
			select {
			case <-guiTick.C:
				g.Update()
			case <-stepTick.C:
				fmt.Println("Stepping")
				if sim.Step() {
					fmt.Println("SOLVED!")
					g.Update()
					return
				}
			}
		}
	}()

	wnd := g.NewMasterWindow("Hello world", 2440, 200, 0, initFont)
	wnd.Run(loop)
}
