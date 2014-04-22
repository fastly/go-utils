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
	all, err := BinaryDuplicateProcessIDs(binary)
	if err != nil {
		return nil, 0, err
	}
	if len(all) > 0 {
		p, err := os.FindProcess(all[0])
		if err != nil {
			return nil, 0, err
		}
		return p, all[0], nil
	}
	return nil, 0, nil
}

// BinaryDuplicateProcessIDs returns all pids belonging to processes with
// the same passed binary name.
func BinaryDuplicateProcessIDs(binary string) (pids []int, err error) {
	infos, err := ioutil.ReadDir("/proc/")
	if err != nil {
		return nil, fmt.Errorf("Couldn't read /proc: %s", err)
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
			pids = append(pids, pid)
		}
	}
	return pids, nil
}

// DuplicateProcessIDs returns all pids belonging to processes with the
// same binary name as the running program.
func DuplicateProcessIDs() (pids []int, err error) {
	binary, err := Path()
	if err != nil {
		return nil, fmt.Errorf("Can't get path: %v", err)
	}
	return BinaryDuplicateProcessIDs(binary)
}
