package executable

// +build linux

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

// Path returns the executable path of the running process.
func Path() (string, error) {
	return os.Readlink("/proc/self/exe")
}

// FindDuplicateProcess looks for any other processes with the same
// binary name as passed in and returns the first one found.
func FindDuplicateProcess(binary string) (*os.Process, int, error) {
	infos, err := ioutil.ReadDir("/proc/")
	if err != nil {
		return nil, 0, fmt.Errorf("Couldn't read /proc: %s", err)
	}
	for _, info := range infos {
		// only want numeric directories
		pid, err := strconv.Atoi(info.Name())
		if err != nil || !info.IsDir() || pid == os.Getpid() {
			continue
		}

		exe, err := os.Readlink("/proc/" + info.Name() + "/exe")
		if err != nil {
			continue
		}
		if exe == binary {
			p, err := os.FindProcess(pid)
			if err != nil {
				return nil, 0, err
			}
			return p, pid, nil
		}
	}
	return nil, 0, nil
}
