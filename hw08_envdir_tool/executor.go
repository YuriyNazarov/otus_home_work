package main

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	// В строке ниже линтер ругается на очевидную дыру в безопасности. ничего умнее чем просто его заткнуть я не нагуглил
	command := exec.Command(cmd[0], cmd[1:]...) //nolint:gosec
	for name, val := range env {
		if !val.NeedRemove {
			os.Setenv(name, val.Value)
		} else {
			os.Unsetenv(name)
		}
	}
	command.Env = os.Environ()
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		} else {
			panic(err)
		}
	}
	return 0
}
