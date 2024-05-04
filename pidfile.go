package main

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
)

var (
	PID_FILE = ".logitter.pid"
)

func init() {
	usr, err := user.Current()
	if err != nil {
		panic(err) // TODO
	}
	PID_FILE = filepath.Join(usr.HomeDir, PID_FILE)
}

func WritePID() {
	pid := strconv.Itoa(os.Getpid())
	err := os.WriteFile(PID_FILE, []byte(pid), 0666)
	if err != nil {
		panic(err) // TODO
	}
}

func PIDExists() bool {
	pid, err := os.ReadFile(PID_FILE)
	if err != nil {
		return false
	}
	intPID, err := strconv.Atoi(string(pid))
	if err != nil {
		return false
	}
	proc, err := os.FindProcess(intPID)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		return false
	}
	return true
}
