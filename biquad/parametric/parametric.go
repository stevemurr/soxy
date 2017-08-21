package parametric

import (
	"math"
	"soxy/biquad"

	"github.com/go-audio/audio"
)

// Parametric eq
type Parametric struct {
	L    biquad.BiQuad
	R    biquad.BiQuad
	Freq float64
	Gain float64
	Q    float64
}

// EQ applies a constant q parametric eq
func EQ(buf *audio.FloatBuffer, freq float64, gain float64, q float64, samplerate float64, channel int) {
	l := Parametric{}
	l.updateCoefficients(samplerate, freq, gain, q)
	for i := 0; i < len(buf.Data); i++ {
		in := buf.Data[i]
		output := l.L.DoBiQuad(in)
		buf.Data[i] = output*l.L.C0 + in*l.L.D0
	}
}

func (p *Parametric) updateCoefficients(samplerate, freq, gain, q float64) {
	K := math.Tan((math.Pi * freq) / samplerate)
	V0 := math.Pow(10, (gain / 20))
	D0 := 1 + ((1 / q) * K) + math.Pow(K, 2)
	E0 := 1 + ((1 / (V0 * q)) * K) + math.Pow(K, 2)
	A := 1 + ((V0 / q) * K) + math.Pow(K, 2)
	B := 2 * (math.Pow(K, 2) - 1)
	G := 1 - ((V0 / q) * K) + math.Pow(K, 2)
	D := 1 - ((1 / q) * K) + math.Pow(K, 2)
	E := 1 - ((1 / (V0 * q)) * K) + math.Pow(K, 2)

	if gain >= 0.0 {
		// Boost
		p.L.A0 = A / D0
		p.L.A1 = B / D0
		p.L.A2 = G / D0
		p.L.B1 = B / D0
		p.L.B2 = D / D0
		p.L.C0 = 1.0
		p.L.D0 = 0.0

		p.R.A0 = A / D0
		p.R.A1 = B / D0
		p.R.A2 = G / D0
		p.R.B1 = B / D0
		p.R.B2 = D / D0
		p.R.C0 = 1.0
		p.R.D0 = 0.0
	} else {
		// Cut
		p.L.A0 = D0 / E0
		p.L.A1 = B / E0
		p.L.A2 = D / E0
		p.L.B1 = B / E0
		p.L.B2 = E / E0
		p.L.C0 = 1.0
		p.L.D0 = 0.0

		p.R.A0 = D0 / E0
		p.R.A1 = B / E0
		p.R.A2 = D / E0
		p.R.B1 = B / E0
		p.R.B2 = E / E0
		p.R.C0 = 1.0
		p.R.D0 = 0.0
	}
}

// func (p *Parametric) updateNotConstantQ(samplerate float64) {
// 	K := (2 * math.Pi * p.Freq) / samplerate
// 	U := math.Pow(10, (p.Gain / 20))
// 	Z := 4 / (1 + U)
// 	B := 0.5 * ((1 - Z*math.Tan(K/(2*p.Q))) / (1 + Z*math.Tan(K/(2*p.Q))))
// 	G := (0.5 + B) * math.Cos(K)

// 	p.L.A0 = 0.5 - B
// 	p.L.A1 = 0.0
// 	p.L.A2 = -(0.5 - B)
// 	p.L.B1 = -2 * G
// 	p.L.B2 = 2 * B
// 	p.L.C0 = U - 1.0
// 	p.L.D0 = 1.0

// 	p.R.A0 = 0.5 - B
// 	p.R.A1 = 0.0
// 	p.R.A2 = -(0.5 - B)
// 	p.R.B1 = -2 * G
// 	p.R.B2 = 2 * B
// 	p.R.C0 = U - 1.0
// 	p.R.D0 = 1.0
// }

// UpdateCoefficients --
// func (p *Parametric) UpdateCoefficients(samplerate float64) {
// 	p.updateConstantQ(samplerate)
// }
