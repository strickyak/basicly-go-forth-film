package film

import (
	"bytes"
	BASIC "github.com/strickyak/basic_basic"
)

func RunBasic(program string) string {
	var buf bytes.Buffer
	putchar := func(ch byte) {
		buf.Write([]byte{ch})
	}
	BASIC.NewTerp(program, putchar).Run()
	return buf.String()
}
