package film

import (
	C "github.com/strickyak/canvas"
	"math/rand"
)

type FloatingTriangle struct {
	X1, Y1, X2, Y2, X3, Y3 int
	R                      *rand.Rand
}

func NewFloatingTriangle(a int) *FloatingTriangle {
	r := NewRand(int64(a))
	return &FloatingTriangle{
		X1: r.Intn(W),
		Y1: r.Intn(H),
		X2: r.Intn(W),
		Y2: r.Intn(H),
		X3: r.Intn(W),
		Y3: r.Intn(H),
		R:  r,
	}
}
func (o *FloatingTriangle) Step(n int) {
	o.X1 += o.R.Intn(2*n+1) - n
	o.Y1 += o.R.Intn(2*n+1) - n
	o.X2 += o.R.Intn(2*n+1) - n
	o.Y2 += o.R.Intn(2*n+1) - n
	o.X3 += o.R.Intn(2*n+1) - n
	o.Y3 += o.R.Intn(2*n+1) - n
}
func (o *FloatingTriangle) Paint(clr C.Color, can *C.Canvas) {
	can.PaintTriangle(o.X1, o.Y1, o.X2, o.Y2, o.X3, o.Y3, clr)
}
