package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args
	envDir := args[1]
	command := args[2:]
	env, err := ReadDir(envDir)
	if err != nil {
		fmt.Println("error occupied while reading env, command not executed", err)
		os.Exit(1)
	}
	os.Exit(RunCmd(command, env))
}
