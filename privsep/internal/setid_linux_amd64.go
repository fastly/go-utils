package internal

// This is a copy of the code removed by
// https://code.google.com/p/go/source/detail?r=ae0d51eadf44. It is regrettably
// required to use the privsep package as of Go 1.4.

import "syscall"

func Setuid(uid int) (err error) {
	_, _, e1 := syscall.RawSyscall(syscall.SYS_SETUID, uintptr(uid), 0, 0)
	if e1 != 0 {
		err = e1
	}
	return
}

func Setgid(gid int) (err error) {
	_, _, e1 := syscall.RawSyscall(syscall.SYS_SETGID, uintptr(gid), 0, 0)
	if e1 != 0 {
		err = e1
	}
	return
}
