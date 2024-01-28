package logger

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

func New(fileName string) *log.Logger {
	f, _ := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	logger := &log.Logger{
		Out:   io.MultiWriter(os.Stderr, f),
		Level: log.TraceLevel,
		Formatter: &myFormatter{log.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        "2006-01-02 15:04:05",
			ForceColors:            true,
			DisableLevelTruncation: true,
		},
		},
	}

	return logger
}
