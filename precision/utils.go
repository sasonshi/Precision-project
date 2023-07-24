package main

import (
	"crypto/sha256"
	"encoding/binary"
	"gonum.org/v1/plot/plotter"
	"math"
	"math/rand"
	"sort"
	"time"
)

type Packet struct {
	flowID         int
	isRecirculated bool
	currStageNum   int
	newVal         int
}

type stage struct {
	keys     []int
	values   []int
	matched  bool
	index    int
	hashFunc func(flowID int) int
}

type pipeline struct {
	keys           []int
	values         []int
	matched        bool
	stages         []stage
	incomingPacket chan Packet
}

func (pip *pipeline) start(isPrecision bool) {
	for packet := range pip.incomingPacket {
		if isPrecision {
			go pip.process(packet)
		} else { // hashpipe
			go pip.processHashPipe(packet)
		}
	}
}

// process in stage
func (stage *stage) process(packet Packet, carryMin int, minStage int) (int, int) {
	l := stage.hashFunc(packet.flowID)
	var oval int

	// hardware stage A
	if l < len(stage.keys) && stage.keys[l] == packet.flowID {
		stage.matched = true
	}

	// hardware stage B
	if stage.matched {
		stage.values[l] = stage.values[l] + 1
	} else {
		oval = stage.values[l]
	}

	// hardware stage C
	if !stage.matched && oval < carryMin {
		carryMin = oval
		minStage = stage.index
	}

	return carryMin, minStage
}

// process in pipeline
func (pip *pipeline) process(packet Packet) {
	firstStage := pip.stages[0]
	carryMin := firstStage.values[firstStage.hashFunc(packet.flowID)] + 1
	minStage := 0
	for i, currStage := range pip.stages {
		if packet.isRecirculated && packet.currStageNum == i {
			l := currStage.hashFunc(packet.flowID)
			currStage.keys[l] = packet.flowID
			currStage.values[l] = packet.newVal
			return
		} else {
			carryMin, minStage = currStage.process(packet, carryMin, minStage)
		}
	}
	isMatched := false
	for _, currStage := range pip.stages {
		if currStage.matched {
			isMatched = true
			currStage.matched = false
		}
	}

	if !isMatched {
		newVal := math.Pow(2, math.Round(math.Log2(float64(carryMin))))
		R := rand.Intn(int(newVal + 1))
		if R == 0 {
			// clone and recirculate packet
			newPacket := clonePacket(packet, minStage, int(newVal))
			pip.incomingPacket <- newPacket
		}
	}
}

func clonePacket(packet Packet, minStage int, newVal int) Packet {
	return Packet{
		flowID:         packet.flowID,
		currStageNum:   minStage,
		isRecirculated: true,
		newVal:         newVal,
	}
}

func hash(flowID int) int {
	xBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(xBytes, uint64(flowID))
	hash := sha256.Sum256(xBytes)
	hashValue := binary.BigEndian.Uint64(hash[:])
	return int(hashValue % (uint64(SRAMPerStage)))
}

func (p *pipeline) sendPackets(flowFrequency []int) {
	clonedFlowFreq := make([]int, len(flowFrequency))
	for i, val := range flowFrequency {
		clonedFlowFreq[i] = val
	}

	sum := 0
	for _, frq := range clonedFlowFreq {
		sum += frq
	}

	for sum > 0 {
		i := rand.Intn(len(clonedFlowFreq))
		for clonedFlowFreq[i] == 0 {
			i = rand.Intn(len(clonedFlowFreq))
		}
		packet := Packet{
			flowID:         i + 1,
			isRecirculated: false,
			currStageNum:   0,
			newVal:         0,
		}
		p.incomingPacket <- packet

		clonedFlowFreq[i] -= 1
		sum--
	}
	time.Sleep(6 * time.Second)
	close(p.incomingPacket)

}

func calculateMSE(array1, array2 []int) float64 {
	if len(array1) != len(array2) {
		panic("Arrays must have the same length")
	}
	packetsNum := 0
	for _, flowCount := range array1 {
		packetsNum += flowCount
	}

	sumSquaredError := 0
	for i := 0; i < len(array1); i++ {
		diff := array1[i] - array2[i]
		squaredError := diff * diff
		sumSquaredError += squaredError
	}

	meanSquaredError := float64(sumSquaredError) / float64(packetsNum)
	return meanSquaredError
}

func topKProblem(estimatedFreq, trueFreq []int, k int) float64 {
	topKIndexesEsimate := findKLargestIndexes(estimatedFreq, k)
	topKIndexesTrue := findKLargestIndexes(trueFreq, k)
	recall := calculateRecallScore(topKIndexesTrue, topKIndexesEsimate, k)
	return recall
}

func findKLargestIndexes(arr []int, k int) []int {
	indexes := make([]int, len(arr))
	for i := 0; i < len(arr); i++ {
		indexes[i] = i
	}

	sort.Slice(indexes, func(i, j int) bool {
		return arr[indexes[i]] > arr[indexes[j]]
	})

	return indexes[:k]
}

func calculateRecallScore(trueIndexes, estimatedIndexes []int, k int) float64 {
	trueSet := make(map[int]bool)
	for _, idx := range trueIndexes {
		trueSet[idx] = true
	}

	hitCount := 0
	for _, idx := range estimatedIndexes {
		if trueSet[idx] {
			hitCount++
		}
	}

	recall := float64(hitCount) / float64(k)
	return recall
}

// generatePoints converts x and y values into plotter.XYs
func generatePoints(xValues, yValues []float64) plotter.XYs {
	points := make(plotter.XYs, len(xValues))
	for i := range points {
		points[i].X = xValues[i]
		points[i].Y = yValues[i]
	}
	return points
}
