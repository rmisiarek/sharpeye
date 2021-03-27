package core

import (
	"fmt"
	"runtime"
)

var (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	cyan   = "\033[36m"
	gray   = "\033[37m"
	white  = "\033[97m"
)

func Green(txt string) string {
	return fmt.Sprintf("%v%v%v", green, txt, reset)
}

func Red(txt string) string {
	return fmt.Sprintf("%v%v%v", red, txt, reset)
}

func Gray(txt string) string {
	return fmt.Sprintf("%v%v%v", gray, txt, reset)
}

func init() {
	if runtime.GOOS == "windows" {
		reset = ""
		red = ""
		green = ""
		yellow = ""
		blue = ""
		purple = ""
		cyan = ""
		gray = ""
		white = ""
	}
}
