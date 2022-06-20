package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

const bufSize = 2048

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	envFiles, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	env := make(Environment)
	for _, file := range envFiles {
		file, err := os.Open(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}
		defer file.Close()
		fileInfo, err := file.Stat()
		if err != nil {
			return nil, err
		}
		if fileInfo.Size() == 0 {
			env[fileInfo.Name()] = EnvValue{
				Value:      "",
				NeedRemove: true,
			}
			continue
		}
		contents, err := getFirstLine(file, fileInfo.Size())
		if err != nil {
			return nil, err
		}
		env[fileInfo.Name()] = EnvValue{
			Value:      contents,
			NeedRemove: false,
		}
	}

	return env, nil
}

func readLine(file *os.File, fileSize int64) ([]byte, error) {
	var readSize, lineSize int64
	buf := make([]byte, bufSize)
	result := buf
	for readSize < fileSize {
		read, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				return result[:lineSize], nil
			}
			return nil, err
		}
		for i := 0; i < read; i++ {
			if buf[i] == '\n' {
				return result[:lineSize], nil
			}
			result = append(result, buf[i])
			lineSize++
		}
		readSize += int64(read)
		buf = make([]byte, bufSize)
	}
	return result[:lineSize], nil
}

func getFirstLine(file *os.File, fileSize int64) (string, error) {
	line, err := readLine(file, fileSize)
	if err != nil {
		return "", err
	}
	lineStr := string(line)
	lineStr = strings.ReplaceAll(lineStr, string([]byte{0}), "\n")
	lineStr = strings.TrimRight(lineStr, " \t")
	return lineStr, nil
}
