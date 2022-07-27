package logger

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var destination = "test_log.log"

func TestLogger(t *testing.T) {
	t.Run("main logger test", func(t *testing.T) {
		defer os.Remove(destination)
		logger := NewLogger("warn", destination)
		// Проверка что файл вывода создан
		_, err := os.Open(destination)
		require.NoError(t, err)
		logger.Debug("debug")
		logger.Info("info")
		logger.Warn("warning")
		logger.Error("error")

		file, err := os.Open(destination)
		require.NoError(t, err)
		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		require.Equal(t, 2, len(lines)) // Должны быть записаны логи уровня >= заданного при создании

		err = logger.Close()
		require.NoError(t, err)
	})

	t.Run("server logger test", func(t *testing.T) {
		logger, err := NewServerLogger(destination)
		require.NoError(t, err)
		_, err = os.Open(destination)
		require.NoError(t, err)
		defer os.Remove(destination)
		logger.Info("ip", "method", "path", "protocol", 200, "latency", "userAgent")
		file, err := os.Open(destination)
		require.NoError(t, err)
		result, err := io.ReadAll(file)
		require.NoError(t, err)
		sResult := strings.Fields(string(result))
		require.Equal(t, "ip", sResult[0])
		err = logger.Close()
		require.NoError(t, err)
	})
}
