package film

import (
	"log"
	"strconv"
)

func Atoi(s string) int {
	n, err := strconv.ParseInt(s, 64, 10)
	if err != nil {
		log.Panicf("not an int: %q: %v", s, err)
	}
	return int(n)
}

func Check(err error) {
	if err != nil {
		log.Panicf("Check Fails: %v", err)
	}
}

func Must(b bool) {
	if b {
		log.Panicf("Must Fails")
	}
}
