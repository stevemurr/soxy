package lowshelf

import "audio-effects/filter/biquad"
import "math"

// LowShelf --
type LowShelf struct {
	L    biquad.BiQuad
	R    biquad.BiQuad
	Freq float64
	Gain float64
}

// Process processes a single sample
func (l *LowShelf) Process(in float64, channel int) float64 {
	var output float64
	if channel == 0 {
		output = l.L.DoBiQuad(in)
		output = output*l.L.C0 + in*l.L.D0
		return output
	}
	output = l.R.DoBiQuad(in)
	output = output*l.R.C0 + in*l.R.D0
	return output
}

// UpdateCoefficients --
func (l *LowShelf) UpdateCoefficients(samplerate float64) {
	theta := 2 * math.Pi * l.Freq / samplerate
	u := math.Pow(10, (l.Gain / 20))
	beta := 4.0 / (1.0 + u)
	omega := beta * math.Tan(theta/2.0)
	gamma := (1 - omega) / (1 + omega)

	l.L.A0 = (1 - gamma) / 2.0
	l.L.A1 = (1 - gamma) / 2.0
	l.L.A2 = 0.0
	l.L.B1 = -gamma
	l.L.B2 = 0.0
	l.L.C0 = u - 1.0
	l.L.D0 = 1.0

	l.R.A0 = (1 - gamma) / 2.0
	l.R.A1 = (1 - gamma) / 2.0
	l.R.A2 = 0.0
	l.R.B1 = -gamma
	l.R.B2 = 0.0
	l.R.C0 = u - 1.0
	l.R.D0 = 1.0
}
