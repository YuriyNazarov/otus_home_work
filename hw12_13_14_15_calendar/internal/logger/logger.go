package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Logger struct {
	output io.WriteCloser
	level  int
}

var levels = map[string]int{
	"debug": 0,
	"info":  1,
	"warn":  2,
	"error": 3,
}

func NewLogger(level string, destination string) *Logger {
	logger := Logger{}
	if destination == "STDERR" {
		logger.output = os.Stderr
	} else {
		outFile, err := os.Create(destination)
		if err != nil {
			logger.output = os.Stdout
			fmt.Println("Could not open log file. Logging to STDOUT")
		} else {
			logger.output = outFile
		}
	}
	lvl, ok := levels[level]
	if !ok {
		logger.output.Write([]byte("Could not parse log level, setting to \"error\""))
		lvl = 3
	}
	logger.level = lvl
	return &logger
}

func (l Logger) Info(msg string) {
	msg = "[INFO] " + time.Now().Format(time.RFC3339) + " " + msg + "\n"
	if l.level <= 1 {
		l.output.Write([]byte(msg))
	}
}

func (l Logger) Error(msg string) {
	msg = "[ERROR] " + time.Now().Format(time.RFC3339) + " " + msg + "\n"
	l.output.Write([]byte(msg))
}

func (l Logger) Warn(msg string) {
	msg = "[WARNING] " + time.Now().Format(time.RFC3339) + " " + msg + "\n"
	if l.level <= 2 {
		l.output.Write([]byte(msg))
	}
}

func (l Logger) Debug(msg string) {
	msg = "[DEBUG] " + time.Now().Format(time.RFC3339) + " " + msg + "\n"
	if l.level == 0 {
		l.output.Write([]byte(msg))
	}
}

func (l *Logger) Close() error {

	return l.output.Close()
}
