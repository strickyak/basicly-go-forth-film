/*
BASICLY GO FORTH
mencoder mf://*.png -mf w=1280:h=720:fps=60:type=png -ovc copy -oac copy -o output.avi
*/
package film

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"

	C "github.com/strickyak/canvas"
	F "github.com/strickyak/kiwisdr-qrss-grabber/font5x7"
	forth "github.com/strickyak/meekly-go-forth"
)

var _ = log.Printf

const (
	W = 1280
	H = 720
)

var (
	BLACK       = C.RGB(0, 0, 0)
	LIGHT_GRAY  = C.RGB(50, 50, 50)
	LIGHT_GREEN = C.RGB(10, 150, 10)
	GREEN       = C.RGB(20, 250, 20)
	BLUE        = C.RGB(20, 20, 250)
	YELLOW      = C.RGB(250, 250, 20)
	WHITE       = C.RGB(250, 250, 250)
	ORANGE      = C.RGB(250, 120, 20)
	RED         = C.RGB(250, 0, 0)
)

func NewRand(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

var N = flag.Int("n", 720, "frames to generate")
var HEADER = flag.Bool("h", false, "produce header")
var TAIL = flag.Bool("t", false, "produce trailer")

// const WhenToStartTerm2 = 1000
const WhenToStartTerm2 = 9200
const WhenToStartTerm3 = 9800

var start_term2, start_term3 bool

// const WhenToStartTerm2 = 1200

func Main() {
	flag.Parse()
	Check(os.MkdirAll("tmp", 0777)) // Temp files go in tmp

	if *HEADER {
		CreateSlate()
	}

	can := C.NewCanvas(1280, 720)
	var term1, term2, term3 *Term
	term1 = NewTerm1()
	var t int
	for t = 0; t < *N; t++ {
		fmt.Fprintf(os.Stderr, "[%d] ", t)

		stop := false
		switch {
		case !start_term2 && t < WhenToStartTerm2:
			stop = CreateFrame(t, can, term1)
		case start_term3 || t > WhenToStartTerm3:
			if term3 == nil {
				term3 = NewTerm3()
			}
			stop = CreateFrame(t, can, term3)
		default:
			if term2 == nil {
				term2 = NewTerm2()
			}
			stop = CreateFrame(t, can, term2)
		}
		if stop {
			break
		}
	}

	if *TAIL {
		RunCredits(can)
	}

	log.Printf("ALL DONE t=%d", t)
}

func Scribble(x0, y0 int, s string, mag int, clr C.Color, can *C.Canvas) {
	for i, ch := range s {
		for col := 0; col < 5; col++ {
			for row := 0; row < 7; row++ {
				if F.Pixel(byte(ch), row, col) {
					for a := 0; a < mag; a++ {
						for b := 0; b < mag; b++ {
							can.Set(x0+mag*(7*i+col)+a, y0-mag*row-b, clr)
						}
					}
				}
			}
		}
	}
}

var ftv = []*FloatingTriangle{
	NewFloatingTriangle(1),
	NewFloatingTriangle(2),
	NewFloatingTriangle(3),
	NewFloatingTriangle(4),
	NewFloatingTriangle(5),
	NewFloatingTriangle(6),
	NewFloatingTriangle(7),
	NewFloatingTriangle(8),
}

var pt = NewPeriTriangle()

const prog3 = `1 REM primes
10 FOR i = 2 TO 5000
20 FOR j = 2 TO i-1
30 IF i % j == 0 THEN 99
40 NEXT j
50 PRINT j
99 NEXT i
`

var out3 = RunBasic(prog3)

// var sc3 = NewScreen(40, 20, 2, W/2, H/2)
var sc3 = NewScreen(32, 40, 3, 0, 0)

// var ks3 = NewKeystroke(prog3 + "RUN\n" + out3)
var ks3 = NewKeystroke(out3)

const prog4 = `1 REM fibonacci
5 LET a = 0
6 LET b = 1
10 FOR i = 2 TO 1000
20 PRINT b
30 LET t = a + b
40 LET a = b
50 LET b = t
99 NEXT i
`
const prog4forth = `variable a
variable b
: fib
    1 b !  0 a !
    0 do
      a @ b @ +
      dup .
        b @ a !
      b !
    loop ;
500 fib
`

//var out4 = RunBasic(prog4)
var out4 = forth.RunForth(prog4forth)

// var sc4 = NewScreen(40, 20, 2, W/2-100, H/20)
// var sc4 = NewScreen(80, 40, 2, W/20, H/20)
var sc4 = NewScreen(32, 42, 3, W/2, 0)

// var ks4 = NewKeystroke(prog4 + "RUN\n" + out4)
var ks4 = NewKeystroke(out4)

var life = NewLife(200, 200, []Point{ // "r" pentomino
	Point{50, 50},
	Point{51, 50},
	Point{49, 51},
	Point{50, 51},
	Point{50, 52},
})

var freeze_tty bool

var random = NewRand(99)

func CreateFrame(t int, can *C.Canvas, term *Term) bool {
	// Wipe a random central section with BLACK.
	can.Fill(random.Intn(W/2), random.Intn(H/2), W/2+random.Intn(W/2), H/2+random.Intn(H/2), BLACK)

	pt.Step(1)
	pt.Paint(LIGHT_GREEN, can)

	for i := 0; i < len(ftv); i++ {
		ftv[i].Step(10)
		ftv[i].Paint(C.RGB(64+7*byte(len(ftv)-i-1), 10, 64+7*byte(i)), can)
	}

	if prog3ready {
		ch3 := ks3.Step()
		if ch3 > 0 && !freeze_tty {
			sc3.Step(ch3)
		}
		sc3.Paint(ORANGE, can)
	}

	if prog4ready {
		ch4 := ks4.Step()
		if ch4 > 0 && !freeze_tty {
			sc4.Step(ch4)
		}
		sc4.Paint(RED, can)
	}

	please_stop := false
	func() {
		defer func() {
			r := recover()
			if r != nil {
				// Only ignore panic(Stop).
				if stop, ok := r.(Stop); ok {
					log.Printf("Stopped: %q", stop)
					please_stop = true
				} else {
					panic(r)
				}
			}
		}()

		// Paint
		const slowdown = 2
		if (t % slowdown) == 0 {
			var bb bytes.Buffer
			term.Draw(&bb) // this actually moves things.
		}
	}()
	term.Paint(WHITE, can)

	can.Grid(100, LIGHT_GRAY)

	const slowLife = 3
	if t%slowLife == 0 {
		life = life.Step()
	}
	life.Paint(0, 0, RED, can)

	memstats := &runtime.MemStats{}
	runtime.ReadMemStats(memstats)
	Scribble(0, 20, fmt.Sprintf(" (%s) %v", runtime.Version(), *memstats), 1, YELLOW, can)

	filename := fmt.Sprintf("tmp/f-%04d.png", t)
	w, err := os.Create(filename)
	Check(err)
	for mag := 2; mag < 4; mag++ {
		Scribble(100+mag*20, mag*20, filename, mag, YELLOW, can)
	}
	can.WritePng(w)
	err = w.Close()
	Check(err)

	return please_stop
}

func Export(can *C.Canvas, filename string) {
	w, err := os.Create(filename)
	Check(err)
	can.WritePng(w)
	err = w.Close()
	Check(err)
}
