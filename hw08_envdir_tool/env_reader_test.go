package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadDir(t *testing.T) {
	t.Run("env reading", func(t *testing.T) {
		expected := Environment{
			"BAR": EnvValue{
				Value:      "bar",
				NeedRemove: false,
			},
			"EMPTY": EnvValue{
				Value:      "",
				NeedRemove: false,
			},
			"FOO": EnvValue{
				Value:      "   foo\nwith new line",
				NeedRemove: false,
			},
			"HELLO": EnvValue{
				Value:      "\"hello\"",
				NeedRemove: false,
			},
			"UNSET": EnvValue{
				Value:      "",
				NeedRemove: true,
			},
		}
		env, err := ReadDir("testdata/env")
		require.NoError(t, err)
		require.Equal(t, expected, env)
	})
}

func TestReadLine(t *testing.T) {
	t.Run("read single line success", func(t *testing.T) {
		file, err := os.Open("testdata/env/BAR")
		require.NoError(t, err)
		fileInfo, err := file.Stat()
		require.NoError(t, err)
		result, err := getFirstLine(file, fileInfo.Size())
		require.NoError(t, err)
		require.Equal(t, "bar", result)
	})

	t.Run("read single empty line", func(t *testing.T) {
		file, err := os.Open("testdata/env/EMPTY")
		require.NoError(t, err)
		fileInfo, err := file.Stat()
		require.NoError(t, err)
		result, err := getFirstLine(file, fileInfo.Size())
		require.NoError(t, err)
		require.Equal(t, "", result)
	})
}
