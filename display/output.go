package d

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Color int

const MaxLength = 30

const (
	_ Color = 30 + iota
	RED
	GREEN
	YELLOW
	BLUE
	PURPLE
	CYAN
	WHITE
)

func Println(v ...interface{}) { fmt.Fprintln(os.Stderr, v...) }
func Print(v ...interface{})   { fmt.Fprint(os.Stderr, v...) }

func C0(cid Color, s string) string {
	return Clr0(cid) + s + ClrRst()
}
func C(cid Color, s string) string {
	return Clr(cid) + s + ClrRst()
}
func Clr0(cid Color) string {
	if cid == WHITE {
		return "\033[0m"
	}
	return "\033[0;" + strconv.Itoa(int(cid)) + "m"
}
func Clr(cid Color) string {
	if cid == WHITE {
		return "\033[1m"
	}
	return "\033[1;" + strconv.Itoa(int(cid)) + "m"
}
func ClrRst() string {
	return "\033[0m"
}

func StripEsc(s string) string {
	const (
		PLAIN = iota
		ESC
		ESC_END
		CSI_1
		CSI_2
		CSI_END
	)
	var state = PLAIN
	var pt string
	for _, c := range s {
		switch {
		case state == PLAIN && c == 0x1b:
			state = ESC
		case state == ESC && c == '[':
			state = CSI_1
		case state == ESC && c != '[':
			state = ESC_END
		case state == CSI_1 && 0x30 <= c && c <= 0x3f:
			state = CSI_1
		case state == CSI_1 && 0x20 <= c && c <= 0x2f:
			state = CSI_2
		case (state == CSI_1 || state == CSI_2) && 0x40 <= c && c <= 0x7e:
			state = CSI_END
		default:
			state = PLAIN
		}
		if state == PLAIN {
			pt += string(c)
		}
	}
	//   _                        _            _
	//  | \                      | 3*         | 2*
	// PLAIN --1b-> ESC --'['-> CSI_1 --2*-> CSI_2
	//  |           /                         /
	//  |----------/---- 40~7e --------------/
	return pt
}

func EscLen(s string) int {
	plainText := StripEsc(s)
	return len(plainText)
}

func ITEM(s string) {
	l := MaxLength - EscLen(s)
	if l < 0 {
		l = 0
		s = s[:MaxLength-2] + ".."
	}
	Print(strings.Repeat(" ", l) + s + " ")
}

var firstSection = true

func SECTION(s string) {
	if firstSection {
		firstSection = false
	} else {
		Println()
	}
	Println(strings.Repeat(" ", MaxLength+1) + C(WHITE, s))
}

func OK() {
	Println(C(CYAN, "OK"))
}

func FAILED() {
	Println(C(RED, "FAILED"))
}

func FAILED_BECAUSE(s string) {
	Println(C(RED, s))
}

func SKIPPED() {
	Println(C0(WHITE, "--"))
}

func ERR(err error) {
	if err == nil {
		OK()
	} else {
		FAILED_BECAUSE(err.Error())
	}
}

func WARN(err error) {
	if err == nil {
		OK()
	} else {
		Println(C(YELLOW, err.Error()))
	}
}
