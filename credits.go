package film

import (
	"fmt"
	"log"

	C "github.com/strickyak/canvas"
)

const Credits = `
           basicly-go-forth [sic]

          by  strick yak  02020-07

for: https://cyberpunkfestival.com/
     (experimental)

license: CC0 1.0 Universal

video: golang code by strick yak:
  github.com/strickyak/basicly-go-forth-film
    basic: github.com/strickyak/basic_basic
    forth: github.com/strickyak/meekly-go-forth

audio: found at https://2600.com/offthehook/
  off_the_hook__19970107  11:24
  off_the_hook__19970218  21:18
original soundtrack by strick yak on
  jerboa modular synthesizer: http://wiki.yak.net/1132
also a few words from Neuromancer by W Gibson

compositor: cujo bird
$ mencoder mf://*.png -mf w=1280:h=720:fps=30:type=png \
    -ovc copy -oac copy -o film.avi
$ ffmpeg  -i film.avi  -i basicly-1511.wav  film.mp4
`

func RunCredits(can *C.Canvas) {
	scr := NewScreen(W, H, 3, 0, 0)
	for i, ch := range Credits + "        " {
		scr.Step(byte(ch))
		scr.OptionalClearUntil(BLACK, can)
		scr.Paint(GREEN, can)
		DrawMorse(can)

		if (i % 2) == 0 {
			filename := fmt.Sprintf("tmp/z-%04d.png", i/2)
			Export(can, filename)
			log.Print(filename)
		}
	}
	can.Fill(0, 0, W, H, BLACK) // finish clearing bg
	scr.Paint(GREEN, can)       // and repaint
	DrawMorse(can)
	for i := 0; i < 90; i++ { // three seconds
		filename := fmt.Sprintf("tmp/z-%04d.png", i+1000)
		Export(can, filename)
		log.Print(filename)
	}
}

func DrawMorse(can *C.Canvas) {
	const Y = 10
	x := 10
	fist := func(b bool) {
		if b {
			can.Set(x, Y, GREEN)
			can.Set(x+1, Y, GREEN)
			can.Set(x, Y-1, GREEN)
			can.Set(x+1, Y-1, GREEN)
		}
		x += 2
	}

	for _, morse := range BAMA {
		switch morse {
		case '.':
			fist(true)
		case '-':
			fist(true)
			fist(true)
			fist(true)
		default:
			fist(false)
		}
		fist(false)
	}
}

const BAMA = ".... --- -- .   .-- .- ...   -... .- -- .- .-.-.-   .... --- -- .   .-- .- ...   - .... .   ... .--. .-. .- .-- .-.. .-.-.-   .... --- -- .   .-- .- ...   - .... .   -... --- ... - --- -.   .- - .-.. .- -. - .-   -- . - .-. --- .--. --- .-.. .. - .- -.   .- -..- .. ... .-.-.- "
