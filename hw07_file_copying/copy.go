package main

import (
	"errors"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

const bufSize int64 = 64

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	inputFile, outFile, err := openFiles(fromPath, toPath)
	if err != nil {
		return err
	}
	defer inputFile.Close()
	defer outFile.Close()

	info, err := inputFile.Stat()
	if err != nil {
		return ErrUnsupportedFile
	}
	if offset > info.Size() {
		return ErrOffsetExceedsFileSize
	}
	if limit == 0 || limit > info.Size()-offset {
		limit = info.Size() - offset
	}
	chunksNum := int(limit / bufSize)
	if limit%bufSize != 0 {
		chunksNum++
	}
	lastChunkSize := limit % bufSize

	inputFile.Seek(offset, 0)
	buf := make([]byte, bufSize)

	var readLen int64
	writeLen := bufSize
	bar := pb.New(chunksNum)
	bar.SetWriter(os.Stdout)
	bar.Start()
	for chunk := 0; chunk < chunksNum; chunk++ {
		readLen = 0
		for readLen < bufSize {
			read, err := inputFile.Read(buf[readLen:])
			readLen += int64(read)
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
		}
		if chunk+1 == chunksNum {
			writeLen = lastChunkSize
		}
		outFile.Write(buf[:writeLen])
		buf = make([]byte, bufSize)
		bar.Increment()
	}
	bar.Finish()
	return nil
}

func openFiles(from, to string) (*os.File, *os.File, error) {
	inputFile, err := os.Open(from)
	if err != nil {
		return nil, nil, err
	}
	outFile, err := os.Create(to)
	if err != nil {
		return nil, nil, err
	}
	return inputFile, outFile, nil
}
