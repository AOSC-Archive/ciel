package d

import (
	"bufio"
	"os"
	"strings"
)

func ASK(msg string, options string) string {
	Println()
	Println(strings.Repeat(" ", MaxLength-3) + C(WHITE, "=== "+msg+" ==="))
	Print(strings.Repeat(" ", MaxLength-3) + ">>> (" + options + "): " + Clr(YELLOW))
	buf := bufio.NewReader(os.Stdin)

	answer, _, _ := buf.ReadLine()
	Println(ClrRst())
	return strings.TrimSpace(string(answer))
}

func ASKLower(msg string, options string) string {
	return strings.ToLower(ASK(msg, options))
}
