package sharpeye

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
)

type infoWriter struct{}
type debugWriter struct{}
type errorWriter struct{}

var (
	infoLogger  *log.Logger = log.New(new(infoWriter), "", 0)
	debugLogger *log.Logger = log.New(new(debugWriter), "", 0)
	errorLogger *log.Logger = log.New(new(errorWriter), "", 0)
)

func (writer infoWriter) Write(bytes []byte) (int, error) {
	return fmt.Printf("%s | %s | %s", time.Now().UTC().Format("15:04:05"), color.BlueString("INFO"), string(bytes))
}

func (writer debugWriter) Write(bytes []byte) (int, error) {
	return fmt.Printf("%s | %s | %s", time.Now().UTC().Format("15:04:05"), color.YellowString("DEBUG"), string(bytes))
}

func (writer errorWriter) Write(bytes []byte) (int, error) {
	return fmt.Printf("%s | %s | %s", time.Now().UTC().Format("15:04:05"), color.RedString("ERROR"), string(bytes))
}

func Info(format string, v ...any) {
	infoLogger.Printf(format, v...)
}

func Debug(format string, v ...any) {
	debugLogger.Printf(format, v...)
}

func Error(format string, v ...any) {
	errorLogger.Printf(format, v...)
}
