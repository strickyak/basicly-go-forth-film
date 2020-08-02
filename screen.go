/*
BASICLY GO FORTH
the STORY of how
a GEEK with no artistic skills
CODES his way through LIFE
mencoder mf://*.png -mf w=1280:h=720:fps=60:type=png -ovc copy -oac copy -o output.avi
*/
package film

import (
	"log"

	C "github.com/strickyak/canvas"
)

// Screen is a rectangular ASCII console you can type on.
type Screen struct {
	Slot   []byte
	W, H   int
	P      int
	Mag    int
	X0, Y0 int
}

func NewScreen(w, h, mag, x0, y0 int) *Screen {
	z := &Screen{
		Slot: make([]byte, w*h),
		W:    w,
		H:    h,
		P:    0,
		Mag:  mag,
		X0:   x0,
		Y0:   y0,
	}
	for i := 0; i < (w * h); i++ {
		z.Slot[i] = 0
	}
	return z
}
func (o *Screen) Step(ch byte) {
	switch ch {
	case '\n':
		for o.P%o.W != o.W-1 {
			o.P++
		}
		o.P++
	default:
		o.Slot[o.P] = ch
		o.P++
	}
	o.P = o.P % (o.W * o.H)
	if o.P%o.W == 0 {
		for i := 0; i < o.W; i++ {
			o.Slot[o.P+i] = 0
		}
	}
}
func (o *Screen) OptionalClearUntil(clr C.Color, can *C.Canvas) {
	min_y := H
	for i := 0; i < o.W; i++ {
		for j := 0; j < o.H; j++ {
			ch := o.Slot[j*o.W+i]
			if ch > 32 {
				y := H - o.Y0 - (j+1)*o.Mag*8
				if y < min_y {
					min_y = y
				}
			}
		}
	}
	can.Fill(0, min_y-(16*o.Mag), W, H, clr)
	log.Printf("min_y = %d", min_y)
}

func (o *Screen) Paint(clr C.Color, can *C.Canvas) {
	for i := 0; i < o.W; i++ {
		for j := 0; j < o.H; j++ {
			ch := o.Slot[j*o.W+i]
			if ch > 32 {
				Scribble(
					o.X0+(i+1)*o.Mag*7,
					H-o.Y0-(j+1)*o.Mag*8,
					string([]byte{ch}), o.Mag, clr, can)
			}
		}
	}
}
