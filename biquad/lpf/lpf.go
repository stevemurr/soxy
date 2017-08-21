package lpf

import (
	"math"
	"soxy/biquad"

	"github.com/go-audio/audio"
)

// LPF implements a butterworth low pass filter
type LPF struct {
	L    biquad.BiQuad
	R    biquad.BiQuad
	Freq float64
}

// LowPass applies a butterworth low pass filter
func LowPass(buf *audio.FloatBuffer, freq float64, samplerate float64, channel int) {
	l := LPF{}
	l.updateCoefficients(samplerate, freq)
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
func (l *LPF) updateCoefficients(samplerate, freq float64) {
	C := 1 / math.Tan(freq/samplerate)
	l.L.A0 = 1 / (1 + math.Sqrt(2)*C + math.Pow(C, 2))
	l.L.A1 = 2 * l.L.A0
	l.L.A2 = l.L.A0
	l.L.B1 = 2 * l.L.A0 * (1 - math.Pow(C, 2))
	l.L.B2 = l.L.A0 * (1 - math.Sqrt(2)*C + math.Pow(C, 2))

	l.L.C0 = 1.0
	l.L.D0 = 0.0

	l.R.A0 = 1 / (1 + math.Sqrt(2)*C + math.Pow(C, 2))
	l.R.A1 = 2 * l.L.A0
	l.R.A2 = l.L.A0
	l.R.B1 = 2 * l.L.A0 * (1 - math.Pow(C, 2))
	l.R.B2 = l.L.A0 * (1 - math.Sqrt(2)*C + math.Pow(C, 2))

	l.R.C0 = 1.0
	l.R.D0 = 0.0
}
