package executable

// +build darwin

// #import <mach-o/dyld.h>
import "C"

import (
	"errors"
	"os"
)

// documentation in executable_linux.go

func Path() (string, error) {
	var buflen C.uint32_t = 1024
	buf := make([]C.char, buflen)

	ret := C._NSGetExecutablePath(&buf[0], &buflen)
	if ret == -1 {
		buf = make([]C.char, buflen)
		C._NSGetExecutablePath(&buf[0], &buflen)
	}
	return C.GoStringN(&buf[0], C.int(buflen)), nil
}

func FindDuplicateProcess(binary string) (*os.Process, int, error) {
	return nil, 0, errors.New("FindDuplicateProcess unimplemented on Darwin")
}

func BinaryDuplicateProcessIDs(binary string) (pids []int, err error) {
	return nil, errors.New("DuplicateProcessIDs unimplemented on Darwin")
}

func DuplicateProcessIDs() (pids []int, err error) {
	return BinaryDuplicateProcessIDs("")
}
