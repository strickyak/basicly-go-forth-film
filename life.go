package film

import (
	C "github.com/strickyak/canvas"
)

type Life struct {
	Slot []byte
	W, H int
}

func NewLife(w, h int, initial []Point) *Life {
	z := &Life{
		Slot: make([]byte, w*h),
		W:    w,
		H:    h,
	}
	for _, p := range initial {
		z.Slot[p.X+w*p.Y] = 1
	}
	return z
}

func (o *Life) Step() *Life {
	z := &Life{
		Slot: make([]byte, o.W*o.H),
		W:    o.W,
		H:    o.H,
	}

	for b := 1; b < o.H-1; b++ {
		for a := 1; a < o.W-1; a++ {
			var n byte
			for i := -1; i < 2; i++ {
				for j := -1; j < 2; j++ {
					n += o.Slot[(a+i)+(b+j)*o.W]
				}
			}
			if o.Slot[a+b*o.W] > 0 {
				if n == 3 || n == 4 {
					z.Slot[a+b*o.W] = 1
				}
			} else {
				if n == 3 {
					z.Slot[a+b*o.W] = 1
				}
			}
		}
	}

	return z
}
func (o *Life) Paint(x, y int, clr C.Color, can *C.Canvas) {
	const mag = 5
	for b := 1; b < o.H-1; b++ {
		for a := 1; a < o.W-1; a++ {
			if o.Slot[a+b*o.W] > 0 {
				for i := 0; i < mag; i++ {
					for j := 0; j < mag; j++ {
						can.Set(i+x+mag*a, j+y+mag*b, clr)
					}
				}
			}
		}
	}
}
