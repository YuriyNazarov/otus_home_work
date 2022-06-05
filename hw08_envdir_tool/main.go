package main

import (
	"os"
)

func main() {
	args := os.Args
	envDir := args[1]
	command := args[2:]
	env, err := ReadDir(envDir)
	if err != nil {
		panic(err)
	}
	os.Exit(RunCmd(command, env))
}
