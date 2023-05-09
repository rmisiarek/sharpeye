package sharpeye

import (
	"strconv"

	"github.com/fatih/color"
)

func Contains(s []int, c int) bool {
	for _, v := range s {
		if v == c {
			return true
		}
	}

	return false
}

func ColorStatus(code int) string {
	switch {
	case code >= 100 && code <= 199:
		return color.BlueString(strconv.Itoa(code))
	case code >= 200 && code <= 299:
		return color.GreenString(strconv.Itoa(code))
	case code >= 300 && code <= 399:
		return color.BlueString(strconv.Itoa(code))
	case code >= 400 && code <= 499:
		return color.YellowString(strconv.Itoa(code))
	case code >= 500 && code <= 599:
		return color.RedString(strconv.Itoa(code))
	default:
		return strconv.Itoa(code)
	}
}
