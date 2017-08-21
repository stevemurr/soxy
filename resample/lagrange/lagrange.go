package lagrange

import (
	"github.com/go-audio/audio"
)

// Resample uses lagrange resampling.  This algorithm proved to be too low quality in the mid range for
// downsampling.
func Resample(buf *audio.FloatBuffer, quality int, old int, new int) *audio.FloatBuffer {
	ratio := float64(old) / float64(new)
	pts := make([]point, quality*2)
	resampled := &audio.FloatBuffer{Format: buf.Format}
	for i := 0; ; i++ {
		j := float64(i) * ratio
		if int(j) >= len(buf.Data) {
			break
		}
		for k := range pts {
			l := int(j) + k - len(pts)/2 + 1
			if l >= 0 && l < len(buf.Data) {
				pts[k] = point{X: float64(l), Y: buf.Data[l]}
			} else {
				pts[k] = point{X: float64(l), Y: 0}
			}
		}
		y := lagrange(pts[:], j)
		resampled.Data = append(resampled.Data, y)
	}
	return resampled
}

func lagrange(pts []point, x float64) (y float64) {
	y = 0.0
	for j := range pts {
		l := 1.0
		for m := range pts {
			if j == m {
				continue
			}
			l *= (x - pts[m].X) / (pts[j].X - pts[m].X)
		}
		y += pts[j].Y * l
	}
	return y
}

type point struct {
	X, Y float64
}
