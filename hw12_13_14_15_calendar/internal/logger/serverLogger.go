package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

type ServerLogger struct {
	output io.WriteCloser
}

func NewServerLogger(destination string) (*ServerLogger, error) {
	logger := ServerLogger{}
	outFile, err := os.Create(destination)
	if err != nil {
		return &logger, fmt.Errorf("could not create file for server logs: %w", err)
	}
	logger.output = outFile
	return &logger, nil
}

func (l ServerLogger) Info(ip, method, path, protocol string, respCode int, latency, userAgent string) {
	msg := fmt.Sprintf("%s [%s] %s %s %s %d %s %s \n",
		ip,
		time.Now().Format("02/Jan/2006:15:04:05 -0700"),
		method,
		path,
		protocol,
		respCode,
		latency,
		userAgent,
	)
	l.output.Write([]byte(msg))
}

func (l *ServerLogger) Close() error {
	return l.output.Close()
}
