package logger

import (
	"fmt"
	"github.com/TwiN/go-color"
	log "github.com/sirupsen/logrus"
	"strings"
)

type myFormatter struct {
	log.TextFormatter
}

func (f *myFormatter) Format(entry *log.Entry) ([]byte, error) {
	// this whole mess of dealing with ansi color codes is required if you want the colored output otherwise you will lose colors in the log levels
	var levelColor int
	switch entry.Level {
	case log.InfoLevel:
		levelColor = 32
	case log.TraceLevel:
		levelColor = 35
	case log.DebugLevel:
		levelColor = 36
	case log.WarnLevel:
		levelColor = 33 // yellow
	case log.ErrorLevel, log.FatalLevel, log.PanicLevel:
		levelColor = 31 // red
	default:
		levelColor = 36 // blue
	}
	//return []byte(fmt.Sprintf("[%s] - \x1b[%dm%s\x1b[0m - %s\n", entry.Time.Format(f.TimestampFormat), levelColor, strings.ToUpper(entry.Level.String()), entry.Message)), nil

	message := fmt.Sprintf(
		"[%s] - \x1b[%dm%s\x1b[0m - %s\n", color.InCyan(entry.Time.Format(f.TimestampFormat)), levelColor, strings.ToUpper(entry.Level.String()), entry.Message)

	for k, v := range entry.Data {
		message += fmt.Sprintf("    %s: %v\n", k, v)
	}

	//message += "\n"

	return []byte(message), nil
}
