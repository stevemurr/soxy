package main

import (
	"audio-effects/compressor"
	"flag"
	"io"
	"log"
	"math"
	"os"

	wav "github.com/youpy/go-wav"
)

var (
	inFile  = flag.String("in", "", "in file")
	outFile = flag.String("out", "", "out file")
)

func main() {
	flag.Parse()

	f, err := os.Open(*inFile)
	if err != nil {
	}

	c := &compressor.Compressor{}
	c.Threshold = -5.0
	c.Ratio = 4.0
	c.AttackTime = 0.1
	c.ReleaseTime = 0.1
	c.OutputGain = 1.0
	c.InputGain = 1.0
	c.SampleRate = 16000.0
	c.LookAheadDelay = 2000
	c.Knee = 20.0
	c.L.DetectMode = 1
	c.R.DetectMode = 1

	c.L.Init(c.SampleRate, c.AttackTime, c.ReleaseTime, false, 2, true)
	c.R.Init(c.SampleRate, c.AttackTime, c.ReleaseTime, false, 2, true)

	c.LDelay.Init(int(0.3 * c.SampleRate))
	c.RDelay.Init(int(0.3 * c.SampleRate))
	c.LDelay.SetDelayInMillis(c.LookAheadDelay)
	c.RDelay.SetDelayInMillis(c.LookAheadDelay)
	w := wav.NewReader(f)
	results := &[]wav.Sample{}
	process(w, c, 16.0, 1, 1.0, results)
	writeWav(*outFile, *results, 16000, 16, 1)
}

func process(w *wav.Reader, c *compressor.Compressor, bitDepth float64, channels uint16, gain float64, results *[]wav.Sample) {
	for {
		samples, err := w.ReadSamples()
		if err == io.EOF {
			break
		}
		for _, sample := range samples {
			y := wav.Sample{}
			y.Values[0] = int(c.Process(w.FloatValue(sample, 0)) * math.Pow(2, 16))
			y.Values[1] = y.Values[0]
			*results = append(*results, y)
		}
	}
}

func writeWav(outFile string, results []wav.Sample, inRate float64, inDepth int, channels uint16) {
	out, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	wr := wav.NewWriter(out, uint32(len(results)), channels, uint32(inRate), uint16(inDepth))
	wr.WriteSamples(results)
}
