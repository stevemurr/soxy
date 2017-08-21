package delay

import (
	"math"
)

// Delay --
type Delay struct {
	DelayInSamples         float64
	OutputAttentuation     float64
	Buffer                 []float64
	ReadIndex              int
	WriteIndex             int
	BufferSize             int
	SampleRate             int
	DelayInMillis          float64
	OutputAttentuationInDB float64
}

func dLinTerp(x1, x2, y1, y2, x float64) float64 {
	denom := x2 - x1
	if denom == 0 {
		return y1 // should not ever happen
	}
	// calculate decimal position of x
	dx := (x - x1) / (x2 - x1)

	// use weighted sum method of interpolating
	result := dx*y2 + (1-dx)*y1

	return result

}

// Init --
func (d *Delay) Init(delayLength int) {
	d.BufferSize = delayLength
	d.Buffer = make([]float64, d.BufferSize)

	// zero out array?
}

// ResetDelay --
func (d *Delay) ResetDelay() {
	// flush buffer

	d.Buffer = make([]float64, d.BufferSize)

	d.WriteIndex = 0
	d.ReadIndex = 0

	d.CookVariables()
}

// SetDelayInMillis -
func (d *Delay) SetDelayInMillis(delayLength float64) {
	d.DelayInMillis = delayLength
	d.CookVariables()
}

// SetOutputAttenuation --
func (d *Delay) SetOutputAttenuation(db float64) {
	d.OutputAttentuationInDB = db
	d.CookVariables()
}

// SetSampleRate --
func (d *Delay) SetSampleRate(sampleRate int) {
	d.SampleRate = sampleRate
	d.CookVariables()
}

// CookVariables --
func (d *Delay) CookVariables() {
	d.OutputAttentuation = math.Pow(10.0, d.OutputAttentuationInDB/20.0)
	d.DelayInSamples = d.DelayInMillis * (float64(d.SampleRate) / 1000.0)
	d.ReadIndex = d.WriteIndex - int(d.DelayInSamples)

	if d.ReadIndex < 0 {
		d.ReadIndex += d.BufferSize
	}
}

// WriteDelayAndInc --
func (d *Delay) WriteDelayAndInc(delayInput float64) {
	d.Buffer[d.WriteIndex] = delayInput
	d.WriteIndex++
	if d.WriteIndex >= d.BufferSize {
		d.WriteIndex = 0
	}
	d.ReadIndex++
	if d.ReadIndex >= d.BufferSize {
		d.ReadIndex = 0
	}
}

// ReadDelay --
func (d *Delay) ReadDelay() float64 {
	// read the output of the delay
	YN := d.Buffer[d.ReadIndex]

	// read the location one behind at y(n-1)
	readIndexBack := d.ReadIndex - 1
	if readIndexBack < 0 {
		readIndexBack = d.BufferSize - 1
	}
	YN1 := d.Buffer[readIndexBack]
	fracDelay := d.DelayInSamples - math.Abs(d.DelayInSamples)
	return dLinTerp(0, 1, YN, YN1, fracDelay) // interp frac between them
}

// ReadDelayAt --
func (d *Delay) ReadDelayAt(secs float64) float64 {

	delayInSamples := secs * float64(d.SampleRate) / 1000.0

	readIndex := d.WriteIndex - int(delayInSamples)

	YN := d.Buffer[readIndex]
	readIndexBack := readIndex - 1
	if readIndexBack < 0 {
		readIndexBack = d.BufferSize - 1
	}
	YN1 := d.Buffer[readIndexBack]
	fracDelay := delayInSamples - math.Abs(delayInSamples)
	return dLinTerp(0, 1, YN, YN1, fracDelay)
}

// ProcessAudio --
func (d *Delay) ProcessAudio(in float64) float64 {
	XN := in
	YN := 0.0
	if d.DelayInSamples == 0 {
		YN = XN
	} else {
		YN = d.ReadDelay()
	}
	d.WriteDelayAndInc(XN)
	return d.OutputAttentuation * YN
}
