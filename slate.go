package film

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/nfnt/resize"
)

/*

https://cyberpunkfestival.com/rules/

The official festival slate sequence must appear at the beginning of all
submissions. It must include 5 seconds of black followed by 7 seconds of
the Cyberpunk Now Film Festival official slate graphic. The Official Slate
Graphic should fill the screen and display the project title and team
member names in the respective fields of the slate in a clear sans-serif
font (such as Arial or Helvetica) with #f0f0f0 light grey text, cutting
to an additional 3 seconds of black following the slate graphic, totaling
15 seconds of Official Slate sequence time at the beginning of the film.

*/

const FPS = 30

const blank1_secs = 5
const slate2_secs = 7
const blank3_secs = 3

func CreateSlate() {
	var slate, blank image.Image
	var err error

	r, err := os.Open("slate.png")
	Check(err)
	slate, err = png.Decode(r)
	Check(err)
	err = r.Close()
	Check(err)
	slate = resize.Resize(W, H, slate, resize.Bilinear)

	blank = image.NewGray16(image.Rectangle{image.Point{0, 0}, image.Point{W, H}})

	WriteHeaderFrames(blank, blank1_secs, 'a')
	WriteHeaderFrames(slate, slate2_secs, 'b')
	WriteHeaderFrames(blank, blank3_secs, 'c')
}

func WriteHeaderFrames(img image.Image, secs int, prefix rune) {
	for i := 0; i < secs*FPS; i++ {
		filename := fmt.Sprintf("tmp/%c-%04d.png", prefix, i)
		w, err := os.Create(filename)
		Check(err)
		png.Encode(w, img)
		w.Close()
		log.Printf("WriteHeadrFrame %q", filename)
	}
}
