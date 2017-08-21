package massberg

import (
	"math"
	"soxy/biquad"

	"github.com/go-audio/audio"
)

// Massberg implements a butterworth low pass filter
type Massberg struct {
	L    biquad.BiQuad
	R    biquad.BiQuad
	Freq float64
	Q    float64
}

// LowPass applies a butterworth low pass filter
func LowPass(buf *audio.FloatBuffer, freq float64, q float64, samplerate float64, channel int) {
	l := Massberg{}
	l.updateCoefficients(samplerate, freq, q)
	for i := 0; i < len(buf.Data); i++ {
		in := buf.Data[i]
		output := l.L.DoBiQuad(in)
		buf.Data[i] = output*l.L.C0 + in*l.L.D0
	}
}

// UpdateCoefficients --
func (l *Massberg) updateCoefficients(samplerate, freq, q float64) {
	theta := 2 * math.Pi * freq / samplerate
	// G1 := 2 / math.Sqrt(2-(math.Pow(math.Sqrt(2)*math.Pi/theta, 2))+(math.Pow(2*math.Pi/q*theta, 2))
	G1 := 2 / math.Sqrt((2)-math.Pow(math.Pow(math.Sqrt(2)*math.Pi/theta, 2), 2)+math.Pow(2*math.Pi/q*theta, 2))
	if q > math.Sqrt(0.5) {
		GR := (2 * math.Pow(q, 2)) / math.Sqrt(4*math.Pow(q, 2)-1)
		WR := theta * math.Sqrt(1-(1/2*math.Pow(q, 2)))
		omegaR := math.Tan(WR / 2)
		omegaS := omegaR * math.Pow((math.Pow(GR, 2)-math.Pow(G1, 2)/math.Pow(GR, 2)-1), 1/4)
		WP := 2 * math.Atan(omegaS)
		WZ := 2 * math.Atan(omegaS/math.Sqrt(G1))
		GP := 1 / math.Sqrt(1-math.Pow(math.Pow(WP/theta, 2), 2)+math.Pow(WP/q*theta, 2))
		GZ := 1 / math.Sqrt(1-math.Pow(math.Pow(WZ/theta, 2), 2)+math.Pow(WZ/q*theta, 2))
		QP := math.Sqrt(G1 * (math.Pow(GP, 2) - math.Pow(GZ, 2)) / (G1 + math.Pow(GZ, 2)*math.Pow(G1-1, 2)))
		QZ := math.Sqrt(math.Pow(G1, 2) * (math.Pow(GP, 2) - math.Pow(GZ, 2)) / math.Pow(GZ, 2) * (G1 + math.Pow(GP, 2)*math.Pow(G1-1, 2)))
		gamma := math.Pow(omegaS, 2) + 1/QP*omegaS + 1
		l.L.A0 = math.Pow(omegaS, 2) + (math.Sqrt(G1)/QZ)*omegaS + G1
		l.L.A1 = 2 * (math.Pow(omegaS, 2) - G1)
		l.L.A2 = math.Pow(omegaS, 2) - (math.Sqrt(G1) / QZ * omegaS) + G1
		l.L.B1 = 2 * (math.Pow(omegaS, 2) - 1)
		l.L.B2 = math.Pow(omegaS, 2) - (1 / QP * omegaS) + 1
		l.L.A0 = l.L.A0 / gamma
		l.L.A1 = l.L.A1 / gamma
		l.L.A2 = l.L.A2 / gamma
		l.L.B1 = l.L.B1 / gamma
		l.L.B2 = l.L.B2 / gamma
		l.L.C0 = 1.0
		l.L.D0 = 0.0
	} else if q <= math.Sqrt(0.5) {

		// GP := 1 / math.Sqrt(1-math.Pow(WP/theta, 2)+math.Pow(WP/q*theta, 2))
		// GZ := 1 / math.Sqrt(1-math.Pow(WZ/theta, 2)+math.Pow(WZ/q*theta, 2))
		// QP := math.Sqrt(G1 * (math.Pow(GP, 2) - math.Pow(G)))
	}
	// C := 1 / math.Tan(freq/samplerate)
	// l.L.A0 = 1 / (1 + math.Sqrt(2)*C + math.Pow(C, 2))
	// l.L.A1 = 2 * l.L.A0
	// l.L.A2 = l.L.A0
	// l.L.B1 = 2 * l.L.A0 * (1 - math.Pow(C, 2))
	// l.L.B2 = l.L.A0 * (1 - math.Sqrt(2)*C + math.Pow(C, 2))

	// l.L.C0 = 1.0
	// l.L.D0 = 0.0

	// l.R.A0 = 1 / (1 + math.Sqrt(2)*C + math.Pow(C, 2))
	// l.R.A1 = 2 * l.L.A0
	// l.R.A2 = l.L.A0
	// l.R.B1 = 2 * l.L.A0 * (1 - math.Pow(C, 2))
	// l.R.B2 = l.L.A0 * (1 - math.Sqrt(2)*C + math.Pow(C, 2))

	// l.R.C0 = 1.0
	// l.R.D0 = 0.0
}
