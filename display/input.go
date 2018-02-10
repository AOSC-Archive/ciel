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
	return strings.ToLower(strings.TrimSpace(string(answer)))
}
