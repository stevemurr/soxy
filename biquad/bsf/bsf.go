package bsf

import (
	"math"
	"soxy/biquad"

	"github.com/go-audio/audio"
)

// BSF implements a butterworth low pass filter
type BSF struct {
	L    biquad.BiQuad
	R    biquad.BiQuad
	Freq float64
}

// BandStop applies a butterworth low pass filter
func BandStop(buf *audio.FloatBuffer, freq float64, samplerate float64, q float64, channel int) {
	l := BSF{}
	l.updateCoefficients(samplerate, freq, q)
	for i := 0; i < len(buf.Data); i++ {
		in := buf.Data[i]
		buf.Data[i] = l.L.DoBiQuad(buf.Data[i])*l.L.C0 + in*l.L.D0
		// if channel == 2 {
		// 	inR := buf.Data[i+1]
		// 	buf.Data[i+1] = l.R.DoBiQuad(buf.Data[i+1])*l.R.C0 + inR*l.R.D0
		// 	i++
		// }
	}
}

// UpdateCoefficients --
func (l *BSF) updateCoefficients(samplerate, freq, q float64) {
	C := math.Tan(math.Pi * freq * (freq / q) / samplerate)
	D := 2 * math.Cos((2*math.Pi*freq)/samplerate)
	l.L.A0 = 1/1 + C
	l.L.A1 = -l.L.A0 * D
	l.L.A2 = l.L.A0
	l.L.B1 = -l.L.A0 * D
	l.L.B2 = l.L.A0 * (1 - C)

	l.L.C0 = 1.0
	l.L.D0 = 0.0

	l.R.A0 = 1/1 + C
	l.R.A1 = -l.L.A0 * D
	l.R.A2 = l.L.A0
	l.R.B1 = -l.L.A0 * D
	l.R.B2 = l.L.A0 * (1 - C)

	l.R.C0 = 1.0
	l.R.D0 = 0.0
}
