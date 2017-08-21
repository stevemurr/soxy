package compressor

import (
	"math"
	"soxy/compressor/delay"
	"soxy/compressor/envelopedetector"

	"github.com/go-audio/audio"
)

// Compressor --
type Compressor struct {
	L envelopedetector.EnvelopeDetector
	R envelopedetector.EnvelopeDetector

	LDelay delay.Delay
	RDelay delay.Delay

	InputGain      float64
	Threshold      float64
	AttackTime     float64
	ReleaseTime    float64
	Ratio          float64
	OutputGain     float64
	Knee           float64
	LookAheadDelay float64
	StereoLink     int
	ProcessorType  int
	SampleRate     float64
	Analog         bool
}

func lagrpol(x []float64, y []float64, n int, xbar float64) float64 {
	fx := 0.0
	l := 1.0
	for i := 0; i < n; i++ {
		l = 1.0
		for j := 0; j < n; j++ {
			if j != i {
				l *= (xbar - x[j]) / (x[i] - x[j])
			}
		}
		fx += l * y[i]
	}
	return fx
}

func (c *Compressor) calcCompressorGain(detectorValue float64, threshold float64, ratio float64, knee float64, limit bool) float64 {
	CS := 1.0 - 1.0/ratio
	if limit {
		CS = 1.0
	}
	if knee > 0 && detectorValue > (threshold-knee/2.0) && detectorValue < threshold+knee/2.0 {
		x := make([]float64, 2)
		y := make([]float64, 2)
		x[0] = threshold - knee/2.0
		x[1] = threshold + knee/2.0
		x[1] = math.Min(0, x[1])
		y[0] = 0
		y[1] = CS
		CS = lagrpol(x, y, 2, detectorValue)
	}
	YG := CS * (threshold - detectorValue)
	YG = math.Min(0, YG)
	return math.Pow(10.0, (YG / 20.0))
}

func multMat(a []float64, b []float64) []float64 {
	results := make([]float64, len(a))
	for idx, el := range a {
		results[idx] = el * b[idx]
	}
	return results
}

// Compress will compress the signal
func Compress(buf *audio.FloatBuffer, ratio float64, attackTime float64, releaseTime float64, threshold float64, inGain float64, outGain float64, sampleRate float64, lookAheadDelay float64, knee float64) {
	/*
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
	*/
	c := Compressor{
		Threshold:      threshold,
		Ratio:          ratio,
		InputGain:      inGain,
		OutputGain:     outGain,
		AttackTime:     attackTime,
		ReleaseTime:    releaseTime,
		SampleRate:     sampleRate,
		LookAheadDelay: lookAheadDelay,
		Knee:           knee,
	}
	c.L.Init(c.SampleRate, c.AttackTime, c.ReleaseTime, false, 2, true)
	c.R.Init(c.SampleRate, c.AttackTime, c.ReleaseTime, false, 2, true)

	c.LDelay.Init(int(0.3 * c.SampleRate))
	c.RDelay.Init(int(0.3 * c.SampleRate))
	c.LDelay.SetDelayInMillis(c.LookAheadDelay)
	c.RDelay.SetDelayInMillis(c.LookAheadDelay)
	for i := 0; i < len(buf.Data); i++ {
		inSample := buf.Data[i]
		inputGain := math.Pow(10.0, c.InputGain/20.0)
		outputGain := math.Pow(10.0, c.OutputGain/20.0)

		XNL := inputGain * inSample

		leftDetector := c.L.Detect(XNL)
		// rightDetector := leftDetector

		linkDetector := leftDetector
		FGN := 1.0

		// linkDetector = 0.5 * (math.Pow(10.0, leftDetector/20.0) + math.Pow(10.0, rightDetector/20.0))
		// linkDetector = 20.0 * math.Log10(linkDetector)

		// set final arg to true to limit
		FGN = c.calcCompressorGain(linkDetector, c.Threshold, c.Ratio, c.Knee, false)
		lookAheadOut := c.LDelay.ProcessAudio(inSample)
		buf.Data[i] = FGN * lookAheadOut * outputGain
	}
}

// Process a sample
func (c *Compressor) Process(inSample float64) float64 {
	inputGain := math.Pow(10.0, c.InputGain/20.0)
	outputGain := math.Pow(10.0, c.OutputGain/20.0)

	XNL := inputGain * inSample

	leftDetector := c.L.Detect(XNL)
	// rightDetector := leftDetector

	linkDetector := leftDetector
	FGN := 1.0

	// linkDetector = 0.5 * (math.Pow(10.0, leftDetector/20.0) + math.Pow(10.0, rightDetector/20.0))
	// linkDetector = 20.0 * math.Log10(linkDetector)

	// set final arg to true to limit
	FGN = c.calcCompressorGain(linkDetector, c.Threshold, c.Ratio, c.Knee, false)
	lookAheadOut := c.LDelay.ProcessAudio(inSample)
	outputSample := FGN * lookAheadOut * outputGain
	return outputSample
}
