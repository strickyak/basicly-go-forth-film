package film

import (
	C "github.com/strickyak/canvas"
)

type Point struct {
	X, Y int
}

var PeriVec = MakePerimeterVec()

func MakePerimeterVec() []Point {
	var z []Point
	for i := 0; i < H; i++ {
		z = append(z, Point{0, i})
	}
	for i := 0; i < W; i++ {
		z = append(z, Point{i, H - 1})
	}
	for i := 0; i < H; i++ {
		z = append(z, Point{W - 1, H - 1 - i})
	}
	for i := 0; i < W; i++ {
		z = append(z, Point{W - 1 - i, 0})
	}
	return z
}

type PeriTriangle struct {
	a, b, c int
}

func NewPeriTriangle() *PeriTriangle {
	return &PeriTriangle{0, H + W/2, H + W + H/2}
}
func (o *PeriTriangle) Step(n int) {
	o.a = (o.a + n) % len(PeriVec)
	o.b = (o.b + n) % len(PeriVec)
	o.c = (o.c + n) % len(PeriVec)
}
func (o *PeriTriangle) Paint(clr C.Color, can *C.Canvas) {
	can.PaintTriangle(PeriVec[o.a].X, PeriVec[o.a].Y, PeriVec[o.b].X, PeriVec[o.b].Y, PeriVec[o.c].X, PeriVec[o.c].Y, clr)
}
