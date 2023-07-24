package main

import (
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"image/color"
	"log"
	"time"
)

func runPrecision(flowFrequency []int) []int {
	values := make([]int, SRAMSize)
	for i := 0; i < len(values); i++ {
		values[i] = initValue
	}
	stages := make([]stage, d)
	pip := pipeline{
		keys:           make([]int, SRAMSize),
		values:         values,
		matched:        false,
		stages:         stages,
		incomingPacket: make(chan Packet),
	}

	for i := 0; i < len(stages); i++ {
		stages[i] = stage{
			keys:     pip.keys[i*SRAMPerStage : (i+1)*SRAMPerStage],
			values:   pip.values[i*SRAMPerStage : (i+1)*SRAMPerStage],
			matched:  false,
			index:    i,
			hashFunc: hash,
		}
	}
	go pip.sendPackets(flowFrequency)

	pip.start(true)
	time.Sleep(time.Second)

	freqCount := make([]int, len(flowFrequency))
	for _, stage := range pip.stages {
		for i, flowID := range stage.keys {
			if flowID > 0 {
				freqCount[flowID-1] += stage.values[i]
			}
		}
	}

	return freqCount
}

// this function creates a graph that shows us how different init values effect the recall score of the top-K problem
func topKPrecision() {
	initValues := []int{2, 3, 5, 10}
	flowFrequency := []int{800, 700, 3, 803, 800, 400, 804, 7, 90000, 350, 900, 720, 17000, 300, 905, 40000, 6, 250, 600, 850, 750, 5, 71, 42, 800,
		40000, 6, 250, 600, 850, 750, 5, 71, 42, 800, 40000, 6, 250, 600, 850, 750, 5, 71, 42, 800}

	numOfCounters := []float64{2, 6, 8, 10}
	k := 15
	// Create a new plot
	p := plot.New()

	// Set plot title and labels
	p.Title.Text = "Curve Plot"
	p.X.Label.Text = "Number Of Counters"
	p.Y.Label.Text = "Recall"

	colors := []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},   // Red
		color.RGBA{R: 0, G: 255, B: 0, A: 255},   // Green
		color.RGBA{R: 0, G: 0, B: 255, A: 255},   // Blue
		color.RGBA{R: 255, G: 255, B: 0, A: 255}, // Yellow
	}
	legendLabels := []string{
		"Init Value 2",
		"Init Value 3",
		"Init Value 5",
		"Init Value 10",
	}

	for i, initVal := range initValues {
		y := make([]float64, len(numOfCounters))
		for i, num := range numOfCounters {
			SRAMSize = d * int(num)
			SRAMPerStage = SRAMSize / d
			initValue = initVal
			precisionFreq := runPrecision(flowFrequency)
			y[i] = topKProblem(precisionFreq, flowFrequency, k)
		}

		// Create a line plot
		line, err := plotter.NewLine(generatePoints(numOfCounters, y))
		if err != nil {
			log.Fatal(err)
		}

		// Set line color
		line.Color = colors[i%len(colors)]

		// Add line plot to the plot
		p.Add(line)

		// Add legend entry
		p.Legend.Add(legendLabels[i], line)
	}

	// Set legend position
	p.Legend.Top = true
	p.Legend.Left = true

	// Save the plot to a file
	err := p.Save(4*vg.Inch, 4*vg.Inch, "curve_plot2.png")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Plot saved to curve_plot2.png")
}

// this function creates a graph that shows us how different init values effect the MSE of the frequency estimator problem
func compareInitValuePrecision() {
	initValues := []int{2, 3, 4}
	flowFrequency := []int{9, 70000, 1, 2, 12000, 38, 2, 25670, 9, 19, 10, 8}
	numOfCounters := []float64{2, 3, 4}

	// Create a new plot
	p := plot.New()

	// Set plot title and labels
	p.Title.Text = "Init Value Precision Comparison"
	p.X.Label.Text = "Num Of Counters"
	p.Y.Label.Text = "MSE"
	colors := []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},   // Red
		color.RGBA{R: 0, G: 255, B: 0, A: 255},   // Green
		color.RGBA{R: 0, G: 0, B: 255, A: 255},   // Blue
		color.RGBA{R: 255, G: 255, B: 0, A: 255}, // Yellow
	}

	legendLabels := []string{
		"Init Value 2",
		"Init Value 3",
		"Init Value 5",
	}

	for i, initVal := range initValues {
		y := make([]float64, len(numOfCounters))
		for i, num := range numOfCounters {
			SRAMSize = d * int(num)
			SRAMPerStage = SRAMSize / d
			initValue = initVal
			precisionFreq := runPrecision(flowFrequency)
			y[i] = calculateMSE(flowFrequency, precisionFreq)
		}

		// Create a line plot
		line, err := plotter.NewLine(generatePoints(numOfCounters, y))
		if err != nil {
			log.Fatal(err)
		}
		// Set line color
		line.Color = colors[i%len(colors)]

		// Add line plot to the plot
		p.Add(line)

		// Add legend entry
		p.Legend.Add(legendLabels[i], line)

	}

	// Set legend position
	p.Legend.Top = true
	p.Legend.Left = true

	// Save the plot to a file
	err := p.Save(4*vg.Inch, 4*vg.Inch, "Init Value Precision Comparison.png")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Init Value Precision Comparison.png")
}

// this function creates a graph that shows us how different values of d effect the MSE of the frequency estimator problem
func compareDSizePrecisionMSE() {
	dValues := []int{1, 2, 3, 4, 5}
	initValue = 5
	flowFrequency := []int{800, 700, 3, 803, 800, 400, 804, 7, 90000, 350, 900, 720, 17000, 300, 905, 40000, 6, 250, 600, 850, 750, 5, 71, 42, 800,
		40000, 6, 250, 600, 850, 750, 5, 71, 42, 800, 40000, 6, 250, 600, 850, 750, 5, 71, 42, 800}

	numOfCounters := []float64{1, 2, 3}

	// Create a new plot
	p := plot.New()

	// Set plot title and labels
	p.Title.Text = "d value PRECISION Comparison"
	p.X.Label.Text = "Number of counters"
	p.Y.Label.Text = "MSE"
	colors := []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},   // Red
		color.RGBA{R: 0, G: 255, B: 0, A: 255},   // Green
		color.RGBA{R: 0, G: 0, B: 255, A: 255},   // Blue
		color.RGBA{R: 255, G: 255, B: 0, A: 255}, // Yellow
		color.RGBA{R: 7, G: 150, B: 120, A: 255}, // I dont know
	}

	legendLabels := []string{
		"d=1",
		"d=2",
		"d=3",
		"d=4",
		"d=5",
	}

	for i, dValue := range dValues {
		y := make([]float64, len(numOfCounters))
		for i, num := range numOfCounters {
			d = dValue
			SRAMSize = d * int(num)
			SRAMPerStage = SRAMSize / d
			precisionFreq := runHashPipe(flowFrequency)
			y[i] = calculateMSE(flowFrequency, precisionFreq)
		}

		// Create a line plot
		line, err := plotter.NewLine(generatePoints(numOfCounters, y))
		if err != nil {
			log.Fatal(err)
		}
		// Set line color
		line.Color = colors[i%len(colors)]

		// Add line plot to the plot
		p.Add(line)

		// Add legend entry
		p.Legend.Add(legendLabels[i], line)

	}

	// Set legend position
	p.Legend.Top = true
	p.Legend.Left = true

	// Save the plot to a file
	err := p.Save(4*vg.Inch, 4*vg.Inch, "d value PRECISION Comparison.png")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("d value PRECISION Comparison.png")
}

// this function creates a graph that shows us how different values of d effect the recall score of the topK problem
func compareDvaluetopKPrecision() {
	initValue = 2
	dValues := []int{1, 2, 3, 4, 5}
	flowFrequency := []int{800, 700, 3, 803, 800, 400, 804, 7, 90000, 350, 900, 720, 17000, 300, 905, 40000, 6, 250, 600, 850, 750, 5, 71, 42, 800,
		40000, 6, 250, 600, 850, 750, 5, 71, 42, 800, 40000, 6, 250, 600, 850, 750, 5, 71, 42, 800}

	numOfCounters := []float64{1, 2, 3}
	k := 15
	// Create a new plot
	p := plot.New()

	// Set plot title and labels
	p.Title.Text = "d value PRECISION topK Comparison"
	p.X.Label.Text = "Number Of Counters"
	p.Y.Label.Text = "Recall"

	colors := []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},   // Red
		color.RGBA{R: 0, G: 255, B: 0, A: 255},   // Green
		color.RGBA{R: 0, G: 0, B: 255, A: 255},   // Blue
		color.RGBA{R: 255, G: 255, B: 0, A: 255}, // Yellow
		color.RGBA{R: 7, G: 150, B: 120, A: 255}, // I dont know

	}

	legendLabels := []string{
		"d=1",
		"d=2",
		"d=3",
		"d=4",
		"d=5",
	}

	for i, dValue := range dValues {
		y := make([]float64, len(numOfCounters))
		for i, num := range numOfCounters {
			d = dValue
			SRAMSize = d * int(num)
			SRAMPerStage = SRAMSize / d
			precisionFreq := runPrecision(flowFrequency)
			y[i] = topKProblem(precisionFreq, flowFrequency, k)
		}

		// Create a line plot
		line, err := plotter.NewLine(generatePoints(numOfCounters, y))
		if err != nil {
			log.Fatal(err)
		}

		// Set line color
		line.Color = colors[i%len(colors)]

		// Add line plot to the plot
		p.Add(line)

		// Add legend entry
		p.Legend.Add(legendLabels[i], line)
	}

	// Set legend position
	p.Legend.Top = true
	p.Legend.Left = true

	// Save the plot to a file
	err := p.Save(4*vg.Inch, 4*vg.Inch, "d value PRECISION Comparison topK.png")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("d value PRECISION Comparison topK.png")
}
