package film

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	C "github.com/strickyak/canvas"
)

type (
	Term struct {
		Mag          int
		TW, TH       int
		S            []byte // screen chars
		Vis          []bool // visibility
		Rooms        []*Room
		Sprites      []*Sprite
		Pipes        []*Pipe
		Monsters     []*Monster
		Darts        []*Dart
		Tick         int
		Tombstone    string
		TombstonePos Point
	}
	Room struct {
		X0, Y0, X9, Y9 int
		Typed          string // chars already typed
		Visible        bool
	}
	Sprite struct {
		Char   byte
		Plan   []Stepper
		RX, RY int // recently at
	}
	Pipe struct {
		Char byte
		Path []Point
	}
	Monster struct {
		Char  byte
		X, Y  int
		Shape []Point // extra shape besides the center point.
	}
	Dart struct {
		X, Y    float64
		DX, DY  float64
		Disable bool
	}
	Stepper interface {
		Step(t *Term, s *Sprite) bool // Return true if this step is done.
	}
	MoveTo struct {
		MX, MY int
	}
	Strafe struct {
		Left, Right, Y int
		Leftward       bool
		Rand           *rand.Rand
	}
	Typewriter struct {
		Room *Room
		S    string
	}
	Stop   string
	Lambda func(t *Term, s *Sprite) bool // Return true if this step is done.
	WaitN  struct {
		N int
	}
	Trigger struct {
		P *bool
	}
)

func NewTerm(tw, th int) *Term {
	return &Term{
		Mag: 3,
		TW:  tw,
		TH:  th,
		S:   make([]byte, tw*th),
		Vis: make([]bool, tw*th),
	}
}

func (t *Term) RenderRoom(r *Room, sprite *Sprite) {
	if r.X0 <= sprite.RX && sprite.RX <= r.X9 && r.Y0 <= sprite.RY && sprite.RY <= r.Y9 {
		r.Visible = true
	}
	if !r.Visible {
		return
	}
	for x := r.X0; x <= r.X9; x++ {
		for y := r.Y0; y <= r.Y9; y++ {
			t.S[x+y*t.TW] = ' '
			t.Vis[x+y*t.TW] = true
		}
	}

	on_a_pipe := func(x, y int) bool {
		for _, pipe := range t.Pipes {
			for _, path := range pipe.Path {
				if path.X == x && path.Y == y {
					return true
				}
			}
		}
		return false
	}
	which_char := func(x, y int, ch byte) byte {
		if on_a_pipe(x, y) {
			return '+'
		}
		return ch
	}
	for y := r.Y0; y <= r.Y9; y++ {
		t.S[r.X0+y*t.TW] = which_char(r.X0, y, '|')
		t.S[r.X9+y*t.TW] = which_char(r.X9, y, '|')
	}
	for x := r.X0; x <= r.X9; x++ {
		t.S[x+r.Y0*t.TW] = which_char(x, r.Y0, '-')
		t.S[x+r.Y9*t.TW] = which_char(x, r.Y9, '-')
	}

	x, y := 1, 1 // cursor at start of paper
	if r.Typed != "" {
		var _x, _y, _i int
		for _, ch := range r.Typed {
			switch ch {
			case '\n':
				x = 1
				y++
			default:
				_x, _y = x+r.X0, y+r.Y0
				_i = _x + _y*t.TW
				t.S[_i] = byte(ch)
				t.Vis[_i] = true
				x++
			}
		}
		_x, _y = x+r.X0, y+r.Y0
		_i = _x + _y*t.TW
		t.S[_i] = sprite.Char
		t.Vis[_i] = true
		if r.Typed[len(r.Typed)-1] != '\177' {
			sprite.RX, sprite.RY = _x, _y
		}
	}
}
func (t *Term) RenderMonster(m *Monster) {
	if m.Char > 0 {
		_j := m.X + m.Y*t.TW
		if 0 <= _j && _j <= len(t.S) {
			t.S[_j] = m.Char
		}
		for _, pt := range m.Shape {
			_i := pt.X + m.X + (pt.Y+m.Y)*t.TW
			if 0 <= _i && _i < len(t.S) { // glitch
				t.S[_i] = m.Char
			}
		}
	}
}

func (trig Trigger) Step(t *Term, s *Sprite) bool {
	*(trig.P) = true
	return true
}
func (w *WaitN) Step(t *Term, s *Sprite) bool {
	w.N--
	return w.N == 0
}
func (s Stop) Step(_ *Term, _ *Sprite) bool {
	log.Printf("Panic(Stop) due to %q", s)
	panic(s)
}
func (tty Typewriter) Step(t *Term, s *Sprite) bool {
	if tty.Room.Typed != tty.S {
		tty.Room.Typed = tty.S[:len(tty.Room.Typed)+1] // one more char
		return false
	}
	return true // finished typing
}

func (o *Strafe) Step(t *Term, s *Sprite) bool {
	if len(t.Monsters) == 0 {
		o.Leftward = false
	}
	for _, monster := range t.Monsters {
		r := o.Rand.Intn(100)
		if r < 33 {
			switch r & 3 {
			case 0:
				monster.X++
			case 1:
				monster.X--
			case 2:
				monster.Y++
			case 3:
				monster.Y--
			}
		}
	}
	if o.Leftward {
		done := MoveTo{o.Left, o.Y}.Step(t, s)
		if done {
			o.Leftward = false
		}
	} else {
		done := MoveTo{o.Right, o.Y}.Step(t, s)
		if done {
			o.Leftward = true
		}
	}
	const slowdown = 3
	if t.Tick&slowdown == 0 && len(t.Monsters) > 0 {
		t.Darts = append(t.Darts, &Dart{
			X:  float64(s.RX),
			Y:  float64(o.Y - 1),
			DX: 0,
			DY: -0.3,
		})
	}
	// When all monsters are gone, we are done.
	return (len(t.Monsters) == 0)
}

func abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}
func near(x1, y1, x2, y2 int) bool {
	return abs(x1-x2) <= 1 && abs(y1-y2) <= 1
}
func dropMonster(vec []*Monster, drop *Monster) []*Monster {
	var z []*Monster
	for _, m := range vec {
		if m != drop {
			z = append(z, m)
		}
	}
	return z
}

func (t *Term) RenderDart(dart *Dart) {
	if dart.Disable {
		return
	}

	dart.X += dart.DX
	dart.Y += dart.DY
	_i := int(dart.X) + int(dart.Y)*t.TW
	if 0 <= _i && _i < len(t.S) {
		t.S[_i] = '|'

		var destroyed *Monster
		for _, m := range t.Monsters {
			if near(m.X, m.Y, int(dart.X), int(dart.Y)) {
				// Dart strikes Monster.
				destroyed = m
				break
			}
		}
		if destroyed != nil {
			t.Monsters = dropMonster(t.Monsters, destroyed)
			dart.Disable = true
		}

	} else {
		dart.Disable = true
	}
}

func (lam Lambda) Step(t *Term, s *Sprite) bool {
	return lam(t, s)
}

func (m MoveTo) Step(t *Term, s *Sprite) bool {
	x, y := s.RX, s.RY
	tx, ty := m.MX, m.MY

	log.Printf("MoveTo: x,y=%d,%d tx,ty=%d,%d", x, y, tx, ty)

	switch { // move one step toward target
	case y < ty: // move in y axis first
		y++
	case y > ty:
		y--
	case x < tx:
		x++
	case x > tx:
		x--
	}

	done := false
	if x == tx && y == ty {
		// arrived at target, remove from plan.
		done = true
	}
	s.RX, s.RY = x, y
	return done
}

func (t *Term) RenderSprite(s *Sprite) {
	// take a step
	if len(s.Plan) > 0 {
		log.Printf("Plan Step %#v", s.Plan[0])
		done := s.Plan[0].Step(t, s)
		if done {
			s.Plan = s.Plan[1:] // remove head
		}
	}
	x, y := s.RX, s.RY
	// show on map
	t.S[x+y*t.TW] = s.Char
	// kill monsters
	for _, m := range t.Monsters {
		if m.X == x && m.Y == y {
			m.Char = 0 // 0 means do not draw
		}
	}
	// make neighbors visible
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			nx, ny := i+x, i+y
			if 0 <= nx && nx < t.TW && 0 <= ny && ny < t.TH {
				t.Vis[nx+ny*t.TW] = true
			}
		}
	}
}
func (t *Term) RenderPipe(p *Pipe) {
	x, y := p.Path[0].X, p.Path[0].Y
	for i := 1; i < len(p.Path); i++ {
		pt := p.Path[i]
		switch {
		case x == pt.X:
			for ; y < pt.Y; y++ {
				t.S[x+y*t.TW] = p.Char
			}
			for ; y > pt.Y; y-- {
				t.S[x+y*t.TW] = p.Char
			}
			t.S[x+y*t.TW] = p.Char
		case y == pt.Y:
			for ; x < pt.X; x++ {
				t.S[x+y*t.TW] = p.Char
			}
			for ; x > pt.X; x-- {
				t.S[x+y*t.TW] = p.Char
			}
			t.S[x+y*t.TW] = p.Char
		}
	}
}

func (t *Term) Draw(w io.Writer) {
	t.Tick++
	scr := make([]byte, len(t.S))
	for i, v := range t.Vis {
		if v {
			scr[i] = ' '
		}
	}
	for _, pipe := range t.Pipes {
		t.RenderPipe(pipe)
	}
	for _, room := range t.Rooms {
		t.RenderRoom(room, t.Sprites[0])
	}
	for _, monster := range t.Monsters {
		t.RenderMonster(monster)
	}
	for _, sprite := range t.Sprites {
		t.RenderSprite(sprite)
	}
	for _, dart := range t.Darts {
		t.RenderDart(dart)
	}
	if enable_tombstone {
		x, y := t.TombstonePos.X, t.TombstonePos.Y
		lines := strings.Split(t.Tombstone, "\n")
		for i, s := range lines {
			for j, ch := range s {
				if ch > 32 {
					_i := x + j + (y+i)*t.TW
					// log.Printf("x=%d y=%d _i=%d", x+j, y+i, _i)
					t.S[_i] = byte(ch)
					t.Vis[_i] = true
				}
			}
		}
	}
	var bb bytes.Buffer
	var k int
	for y := 0; y < t.TH; y++ {
		for x := 0; x < t.TW; x++ {
			ch := t.S[k]
			switch ch {
			case 0:
				ch = ' '
			}
			bb.WriteByte(ch)
			k++
		}
		bb.WriteByte('\n')
	}
	fmt.Fprintf(w, "\n;;;;;;;;;;;;\n%s\n;;;;;;;;;;;;\n", bb.String())
}

func Times(n int, fn func()) {
	for i := 0; i < n; i++ {
		fn()
	}
}

func Move(speed, lx, ly, px, py int, z *[]Point) {
	switch {
	case lx < px:
		for lx < px {
			lx++
			Times(speed, func() {
				*z = append(*z, Point{lx, ly})
			})
		}
	case lx > px:
		for lx > px {
			lx--
			Times(speed, func() {
				*z = append(*z, Point{lx, ly})
			})
		}
	}
	switch {
	case ly < py:
		for ly < py {
			ly++
			Times(speed, func() {
				*z = append(*z, Point{lx, ly})
			})
		}
	case ly > py:
		for ly > py {
			ly--
			Times(speed, func() {
				*z = append(*z, Point{lx, ly})
			})
		}
	}
}

var M_CELL = regexp.MustCompile(`^([a-z])([0-9]+)$`).FindStringSubmatch
var M_WAIT = regexp.MustCompile(`^([-+@])([0-9]+)$`).FindStringSubmatch

func (o *Term) Paint(clr C.Color, can *C.Canvas) {
	const (
		X0 = 0
		Y0 = 50
	)
	for i := 0; i < o.TW; i++ {
		for j := 0; j < o.TH; j++ {
			if o.Vis[j*o.TW+i] {
				ch := o.S[j*o.TW+i]
				if ch > 32 {
					Scribble(
						X0+(i+1)*o.Mag*7,
						H-Y0-(j+1)*o.Mag*8,
						string([]byte{ch}), o.Mag, clr, can)
				}
			}
		}
	}
}

var alienRand = NewRand(2001)

func Alien(x, y int, char byte) *Monster {
	return &Monster{
		X: x, Y: y, Char: char,
		Shape: []Point{
			Point{alienRand.Intn(3) - 1, alienRand.Intn(3) - 1},
			Point{alienRand.Intn(3) - 1, alienRand.Intn(3) - 1},
			Point{alienRand.Intn(3) - 1, alienRand.Intn(3) - 1},
			Point{alienRand.Intn(3) - 1, alienRand.Intn(3) - 1},
			Point{alienRand.Intn(3) - 1, alienRand.Intn(3) - 1},
		},
	}
}

var enable_tombstone bool

func NewTerm3() *Term {
	t := NewTerm(80, 28)
	t.Rooms = []*Room{
		&Room{20, 6, 26, 12, "", false},
	}
	t.Pipes = []*Pipe{
		&Pipe{
			Char: '#',
			Path: []Point{Point{0, 8}, Point{20, 8}},
		},
		&Pipe{ // never uncovered
			Char: '#',
			Path: []Point{Point{25, 10}, Point{30, 10}},
		},
	}
	t.Monsters = []*Monster{
		&Monster{Char: 'E', X: 22, Y: 8}, // Emu of death.
	}
	t.Sprites = []*Sprite{
		&Sprite{
			Char: '@',
			RX:   0, RY: 8,
			Plan: []Stepper{
				MoveTo{0, 8},
				&WaitN{10},
				MoveTo{20, 8},
				&WaitN{10},
				MoveTo{21, 8},
				&WaitN{5},
				Lambda(func(t *Term, s *Sprite) bool {
					t.Monsters[0].X--       // move the Emu
					t.Sprites[0].Char = 'E' // hide the dead the player behind the Emu
					return true
				}),
				&WaitN{5},
				Trigger{&enable_tombstone},
				Trigger{&freeze_tty},
				&WaitN{20},
				Stop("Finished Term3"),
			},
		},
	}
	t.TombstonePos = Point{25, 10}
	t.Tombstone = `
           __________
          /          \
         /    REST    \
        /      IN      \
       /     PEACE      \
      /                  \
      |      strick      |
      |      255 Au      |
      |   killed by an   |
      |       emu        |
      |       2020       |
     *|     *  *  *      | *
_____)/\\_//(\/(/\)/\//\/|_)_____
`
	return t
}

func NewTerm2() *Term {
	const right = 58
	t := NewTerm(right, 24)
	t.Rooms = []*Room{
		&Room{6, 0, right - 5, 22, "", false},
	}
	t.Sprites = []*Sprite{
		&Sprite{
			Char: '@',
			RX:   0, RY: 0,
			Plan: []Stepper{

				MoveTo{0, 0},
				MoveTo{0, 20},
				&WaitN{5},
				MoveTo{6, 20},
				&WaitN{5},

				&Strafe{6, right - 6, 20, false, NewRand(888)},

				MoveTo{right - 6, 20},
				MoveTo{right - 6, 8},
				MoveTo{right, 8},
				&WaitN{10},
				Trigger{&start_term3},
			},
		},
	}
	t.Pipes = []*Pipe{
		&Pipe{
			Char: '#',
			Path: []Point{Point{0, 0}, Point{0, 20}},
		},
		&Pipe{
			Char: '#',
			Path: []Point{Point{0, 20}, Point{6, 20}},
		},
		&Pipe{
			Char: '#',
			Path: []Point{Point{right - 5, 8}, Point{right, 8}},
		},
	}
	var alienChars = []byte("$%^&*+")
	for i := 0; i < len(alienChars); i++ {
		t.Monsters = append(t.Monsters, Alien(10+i*7, 4, alienChars[i]))
	}
	for i := 0; i < len(alienChars); i++ {
		t.Monsters = append(t.Monsters, Alien(9+i*7, 9, alienChars[i]))
	}
	for i := 0; i < len(alienChars); i++ {
		t.Monsters = append(t.Monsters, Alien(11+i*7, 14, alienChars[i]))
	}
	return t
}

var prog3ready bool
var prog4ready bool

func NewTerm1() *Term {
	t := NewTerm(80, 24)
	t.Rooms = []*Room{
		&Room{30, 0, 55, 7, "", false},
		&Room{20, 10, 60, 20, "", false},
		&Room{2, 3, 15, 15, "", false},
	}
	t.Sprites = append(t.Sprites, &Sprite{
		Char: '@',
		RX:   0, RY: 0,
		Plan: []Stepper{

			MoveTo{0, 0},
			MoveTo{12, 0},
			MoveTo{12, 4},
			&WaitN{10},

			MoveTo{3, 4},
			MoveTo{14, 4},
			MoveTo{3, 4},

			MoveTo{5, 11}, // *
			MoveTo{8, 12}, // !
			MoveTo{13, 14},
			MoveTo{22, 14},
			&WaitN{10},
			MoveTo{25, 15}, // B
			MoveTo{58, 18}, // S
			MoveTo{56, 15}, // *
			MoveTo{26, 11}, // *
			MoveTo{22, 12}, // *
			MoveTo{40, 12},
			MoveTo{40, 6},
			&WaitN{10},
			MoveTo{31, 1},
			Typewriter{t.Rooms[0], prog3 + "\nRUN\n\177"},
			Trigger{&prog3ready},
			&WaitN{50},
			MoveTo{21, 11},
			Typewriter{t.Rooms[1], prog4forth + "\n\177"},
			Trigger{&prog4ready},
			&WaitN{50},
			MoveTo{78, 23},
			&WaitN{1},
			Trigger{&start_term2},
		},
	})
	t.Monsters = []*Monster{
		&Monster{Char: '!', X: 5, Y: 11},
		&Monster{Char: '*', X: 8, Y: 12},
		&Monster{Char: 'B', X: 25, Y: 15, Shape: []Point{
			Point{-1, -1},
			Point{-1, +1},
			Point{+1, -1},
			Point{+1, +1},
		}},
		&Monster{Char: 'S', X: 58, Y: 18},
		&Monster{Char: '*', X: 56, Y: 15},
		&Monster{Char: '*', X: 26, Y: 11},
		&Monster{Char: '*', X: 22, Y: 12},
	}
	t.Pipes = []*Pipe{
		&Pipe{
			Char: '#',
			Path: []Point{Point{0, 0}, Point{12, 0}},
		},
		&Pipe{
			Char: '#',
			Path: []Point{Point{12, 0}, Point{12, 3}},
		},
		&Pipe{
			Char: '#',
			Path: []Point{Point{15, 14}, Point{20, 14}},
		},
		&Pipe{
			Char: '#',
			Path: []Point{Point{40, 7}, Point{40, 10}},
		},
	}
	return t
}

func RunTerm() {
	t := NewTerm1()
	for {
		t.Draw(os.Stdout)
		time.Sleep(20 * time.Millisecond)
	}
}
