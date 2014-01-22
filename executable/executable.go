// Package executable has functions to return the executable path
// or directory and other process testing functions.
package executable

import (
	"github.com/fastly/go-utils/vlog"
	"log"
	"path/filepath"
	"syscall"
)

// NowRunning returns true if there is a running process whose
// binary has the same name as this one.
func NowRunning() bool {
	binary, err := Path()
	if err != nil {
		log.Fatalf("Couldn't find own process: %s", err)
		return false
	}
	proc, _, err := FindProcess(binary)
	if err != nil {
		vlog.VLogf("Couldn't look for running processes: %s", err)
		return false
	}
	if proc == nil {
		return false
	}
	if proc.Signal(syscall.Signal(0x0)) == nil {
		return true
	}
	return false
}

// Dir returns the running executable process's directory.
func Dir() (dir string, err error) {
	path, err := Path()
	if err != nil {
		return
	}
	dir, _ = filepath.Split(path)
	return
}
