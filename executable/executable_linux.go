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

// FindProcess returns the running process and it's PID given by the string path
// or an error.
func FindProcess(binary string) (*os.Process, int, error) {
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
