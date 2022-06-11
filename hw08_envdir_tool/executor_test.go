package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	t.Run("execution test", func(t *testing.T) {
		// Создать скрипт выводящий один енв, выходящий с кодом != 0
		outFile, err := os.Create("testdata/code_test.sh")
		require.NoError(t, err)
		script := []byte("#!/usr/bin/env bash\necho -e \"${HELLO}\"\nexit 42")
		_, err = outFile.Write(script)
		require.NoError(t, err)

		// Данные для теста и запуск
		input := []string{
			"/bin/bash",
			"testdata/code_test.sh",
		}
		env, err := ReadDir("testdata/env")
		require.NoError(t, err)
		resultCode := RunCmd(input, env)
		require.Equal(t, 42, resultCode)

		// Очистка
		os.Remove("testdata/code_test.sh")
	})
}
