package film

import (
	"bytes"
	"fmt"
)

// Keystroke is a source of keystrokes, which the Step method returns.
type Keystroke struct {
	Bytes []byte
	Pos   int
}

func NewKeystroke(s string) *Keystroke {
	return &Keystroke{[]byte(s), 0}
}
func (o *Keystroke) Step() byte {
	if o.Pos >= len(o.Bytes) {
		return 0
	}
	z := o.Bytes[o.Pos]
	o.Pos++
	return z
}

// KSSquares is a an instance of Keystroke that emits squares of the naturals.
var KSSquares *Keystroke

func init() {
	var bb bytes.Buffer
	for i := 0; i < 1000; i++ {
		fmt.Fprintf(&bb, "%d ", i*i)
		if (i+1)%8 == 0 {
			fmt.Fprintf(&bb, "\n")
		}
	}
	KSSquares = NewKeystroke(bb.String())
}
