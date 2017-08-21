package biquad

const (
	// FLTEpsilonPlus --
	FLTEpsilonPlus = 1.192092896e-07 /* smallest such that 1.0+FLT_EPSILON != 1.0 */
	// FLTEpsilonMinus --
	FLTEpsilonMinus = -1.192092896e-07 /* smallest such that 1.0-FLT_EPSILON != 1.0 */
	// FLTMinPlus --
	FLTMinPlus = 1.175494351e-38 /* min positive value */
	// FLTMinMinus --
	FLTMinMinus = -1.175494351e-38 /* min negative value */
)

// BiQuad implements a modified biquad filter with wet and dry coefficients.
type BiQuad struct {
	A0 float64
	A1 float64
	A2 float64
	B1 float64
	B2 float64

	// Wet and Dry
	C0 float64
	D0 float64

	// Delays
	XZ1 float64
	XZ2 float64
	YZ1 float64
	YZ2 float64
}

// FlushDelays flushes the delays
func (b *BiQuad) FlushDelays() {
	b.XZ1 = 0
	b.XZ2 = 0
	b.YZ1 = 0
	b.YZ2 = 0
}

// DoBiQuad --
func (b *BiQuad) DoBiQuad(xn float64) float64 {
	// just do the difference equation: y(n) = a0x(n) + a1x(n-1) + a2x(n-2) - b1y(n-1) - b2y(n-2)
	yn := b.A0*xn + b.A1*b.XZ1 + b.A2*b.XZ2 - b.B1*b.YZ1 - b.B2*b.YZ2
	// underflow check
	if yn > 0.0 && yn < FLTMinPlus {
		yn = 0
	}
	if yn < 0.0 && yn > FLTMinMinus {
		yn = 0
	}

	// shuffle delays
	// Y delays
	b.YZ2 = b.YZ1
	b.YZ1 = yn

	// X delays
	b.XZ2 = b.XZ1
	b.XZ1 = xn

	return yn
}
