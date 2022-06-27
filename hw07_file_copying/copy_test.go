package main

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	outPath = "out.txt"
	inPath  = "testdata/input.txt"
)

func TestCopy(t *testing.T) {
	t.Run("file operations success", func(t *testing.T) {
		// Провека успешности окрытия файлов
		_, _, err := openFiles(inPath, outPath)
		require.NoError(t, err)
		// Проверка что файл вывода создан
		_, err = os.Open(outPath)
		require.NoError(t, err)
		os.Remove(outPath)
	})

	t.Run("file operations fail", func(t *testing.T) {
		// Провека успешности окрытия файлов - ошибка
		_, _, err := openFiles(inPath+"asd", outPath)
		require.True(t, os.IsNotExist(err))
		// Провека что при ошибке файл вывода не создается
		_, err = os.Open(outPath)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("copy full file", func(t *testing.T) {
		err := Copy(inPath, outPath, 0, 0)
		require.NoError(t, err)
		inFile, err := os.Open(inPath)
		require.NoError(t, err)
		outFile, err := os.Open(outPath)
		require.NoError(t, err)
		infoIn, _ := inFile.Stat()
		infoOut, _ := outFile.Stat()
		require.Equal(t, infoIn.Size(), infoOut.Size())
		os.Remove(outPath)
	})

	t.Run("copy 1 byte", func(t *testing.T) {
		err := Copy(inPath, outPath, 0, 1)
		require.NoError(t, err)
		outFile, err := os.Open(outPath)
		require.NoError(t, err)
		infoOut, _ := outFile.Stat()
		require.Equal(t, int64(1), infoOut.Size())
		os.Remove(outPath)
	})

	t.Run("copy 5 bytes from offset", func(t *testing.T) {
		expected := "Docum"
		err := Copy(inPath, outPath, 3, 5)
		require.NoError(t, err)
		outFile, err := os.Open(outPath)
		require.NoError(t, err)
		buf := make([]byte, 5)
		num, err := io.ReadFull(outFile, buf)
		require.NoError(t, err)
		require.Equal(t, num, 5)
		require.Equal(t, expected, string(buf))
		os.Remove(outPath)
	})

	t.Run("copy offset more than file", func(t *testing.T) {
		err := Copy(inPath, outPath, 99999999, 1)
		require.Equal(t, err, ErrOffsetExceedsFileSize)
		os.Remove(outPath)
	})

	t.Run("copy limit more than file", func(t *testing.T) {
		err := Copy(inPath, outPath, 0, 999999999)
		require.NoError(t, err)
		inFile, err := os.Open(inPath)
		require.NoError(t, err)
		outFile, err := os.Open(outPath)
		require.NoError(t, err)
		infoIn, _ := inFile.Stat()
		infoOut, _ := outFile.Stat()
		require.Equal(t, infoIn.Size(), infoOut.Size())
		os.Remove(outPath)
	})
}
