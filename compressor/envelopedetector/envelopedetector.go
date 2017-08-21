package envelopedetector

import "math"

const (
	// DigitalTC --
	DigitalTC = -2.0 // log(1%)
	// AnalogTC --
	AnalogTC = -0.43533393574791066201247090699309 // (log(36.7%)
	// FLTEpsilonPlus --
	FLTEpsilonPlus = 1.192092896e-07 /* smallest such that 1.0+FLT_EPSILON != 1.0 */
	// FLTEpsilonMins --
	FLTEpsilonMins = -1.192092896e-07 /* smallest such that 1.0-FLT_EPSILON != 1.0 */
	// FLTMinPlus --
	FLTMinPlus = 1.175494351e-38 /* min positive value */
	// FLTMinMinus --
	FLTMinMinus = -1.175494351e-38 /* min negative value */
)

// EnvelopeDetector --
type EnvelopeDetector struct {
	AttackTimeInMillis  float64
	ReleastTimeInMillis float64
	AttackTime          float64
	ReleaseTime         float64
	SampleRate          float64
	Envelope            float64
	DetectMode          int
	Sample              int
	AnalogTC            bool
	LogDetector         bool
}

// Init --
func (e *EnvelopeDetector) Init(sampleRate float64, attackInMillis float64, releaseInMillis float64, analog bool, detect int, logDetector bool) {
	e.Envelope = 0.0
	e.SampleRate = sampleRate
	e.AnalogTC = analog
	e.AttackTimeInMillis = attackInMillis
	e.ReleastTimeInMillis = releaseInMillis
	e.DetectMode = detect
	e.LogDetector = logDetector

	e.setAttackTime(attackInMillis)
	e.setReleaseTime(releaseInMillis)
}

func (e *EnvelopeDetector) setAttackTime(attackInMillis float64) {
	e.AttackTime = attackInMillis

	if e.AnalogTC {
		e.AttackTime = math.Exp(AnalogTC / (attackInMillis * e.SampleRate * 0.001))
	} else {
		e.AttackTime = math.Exp(DigitalTC / (attackInMillis * e.SampleRate * 0.001))
	}
}

func (e *EnvelopeDetector) setReleaseTime(releaseInMillis float64) {
	e.ReleastTimeInMillis = releaseInMillis

	if e.AnalogTC {
		e.ReleaseTime = math.Exp(AnalogTC / (releaseInMillis * e.SampleRate * 0.001))
	} else {
		e.ReleaseTime = math.Exp(DigitalTC / (releaseInMillis * e.SampleRate * 0.001))
	}
}

func (e *EnvelopeDetector) setTCModeAnalog(analogTC bool) {
	e.AnalogTC = analogTC
	e.setAttackTime(e.AttackTimeInMillis)
	e.setReleaseTime(e.ReleastTimeInMillis)
}

// Detect --
func (e *EnvelopeDetector) Detect(input float64) float64 {
	switch e.DetectMode {
	case 0:
		input = math.Abs(input)
		break
	case 1:
		input = math.Abs(input) * math.Abs(input)
		break
	case 2:
		input = math.Pow(math.Abs(input)*math.Abs(input), 0.5)
		break
	default:
		input = math.Abs(input)
		break
	}

	// var old float64 = e.Envelope
	if input > e.Envelope {
		e.Envelope = e.AttackTime*(e.Envelope-input) + input
	} else {
		e.Envelope = e.ReleaseTime*(e.Envelope-input) + input
	}
	if e.Envelope > 0.0 && e.Envelope < FLTMinPlus {
		e.Envelope = 0
	}
	if e.Envelope < 0.0 && e.Envelope > FLTMinMinus {
		e.Envelope = 0
	}
	e.Envelope = math.Min(e.Envelope, 1.0)
	e.Envelope = math.Max(e.Envelope, 0.0)

	if e.LogDetector {
		if e.Envelope <= 0 {
			return -96.0
		}
		return 20 * math.Log10(e.Envelope)
	}
	return e.Envelope
}
