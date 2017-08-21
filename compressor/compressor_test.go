package compressor

import (
	"io"
	"log"
	"math"
	"os"
	"testing"

	wav "github.com/youpy/go-wav"
)

func TestCompressor(t *testing.T) {

	f, err := os.Open("test.wav")
	if err != nil {
		t.Fatal("cant open test wav")
	}

	c := &Compressor{}
	c.Threshold = -20.0
	c.Ratio = 10.0
	c.AttackTime = 30.0
	c.ReleaseTime = 1000.0
	c.OutputGain = 0.0
	c.InputGain = 0.0
	c.SampleRate = 16000.0
	c.L.DetectMode = 2
	c.R.DetectMode = 2
	c.L.AnalogTC = true

	c.Init()
	w := wav.NewReader(f)
	results := &[]wav.Sample{}
	process(w, c, 16.0, 1, 1.0, results)
	writeWav("out.wav", *results, 16000, 16, 1)
}

// func processSample(in float64, , bitDepth float64, channel int) float64 {
// 	for _, filt := range fx {
// 		in = filt.Process(in, channel)
// 	}
// 	return in
// }

func process(w *wav.Reader, c *Compressor, bitDepth float64, channels uint16, gain float64, results *[]wav.Sample) {
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
