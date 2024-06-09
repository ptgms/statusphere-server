package main

import (
	"bytes"
	"github.com/shirou/gopsutil/v3/process"
	"os/exec"
)

// verifyRoot checks if the given password is correct for the root user.
func verifyRoot(password string) bool {
	// prepare the sudo command
	cmd := exec.Command("sudo", "-S", "echo", "root_check")

	// create a buffer to hold the password input
	var stdin bytes.Buffer
	stdin.Write([]byte(password + "\n"))
	cmd.Stdin = &stdin

	// create a buffer to capture the command output
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// run the command
	err := cmd.Run()

	// check the output and error
	if err != nil {
		return false
	}
	return stdout.String() == "root_check\n"
}

func killProcess(pid int) error {
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return err
	}
	return proc.Terminate()
}
