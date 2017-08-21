package hpf

import (
	"math"
	"soxy/biquad"

	"github.com/go-audio/audio"
)

// HPF --
type HPF struct {
	L    biquad.BiQuad
	R    biquad.BiQuad
	Freq float64
}

// HighPass applies a butterworth high pass filter
func HighPass(buf *audio.FloatBuffer, freq float64, samplerate float64, channel int) {
	l := HPF{}
	l.updateCoefficients(samplerate, freq)
	for i := 0; i < len(buf.Data); i++ {
		in := buf.Data[i]
		output := l.L.DoBiQuad(in)
		buf.Data[i] = output*l.L.C0 + in*l.L.D0
	}
}

// UpdateCoefficients --
func (l *HPF) updateCoefficients(samplerate, freq float64) {
	C := math.Tan(freq / samplerate)
	l.L.A0 = 1 / (1 + math.Sqrt(2)*C + math.Pow(C, 2))
	l.L.A1 = -2 * l.L.A0
	l.L.A2 = l.L.A0
	l.L.B1 = 2 * l.L.A0 * (math.Pow(C, 2) - 1)
	l.L.B2 = l.L.A0 * (1 - math.Sqrt(2)*C + math.Pow(C, 2))

	l.L.C0 = 1.0
	l.L.D0 = 0.0

	l.R.A0 = 1 / (1 + math.Sqrt(2)*C + math.Pow(C, 2))
	l.R.A1 = -2 * l.L.A0
	l.R.A2 = l.L.A0
	l.R.B1 = 2 * l.L.A0 * (math.Pow(C, 2) - 1)
	l.R.B2 = l.L.A0 * (1 - math.Sqrt(2)*C + math.Pow(C, 2))

	l.R.C0 = 1.0
	l.R.D0 = 0.0
}
