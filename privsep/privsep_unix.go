// +build linux

package privsep

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"syscall"

	"github.com/fastly/go-utils/executable"
)

func createChild(username, bin string, args []string) (r io.Reader, w io.Writer, err error) {
	// create a pipe for each direction
	var childIn, childOut, parentIn, parentOut *os.File
	childIn, parentOut, err = os.Pipe()
	if err != nil {
		return
	}
	parentIn, childOut, err = os.Pipe()
	if err != nil {
		return
	}

	child := exec.Command(bin, origArgs[1:]...)

	// childIn becomes fd 3 in child, childOut becomes fd 4
	child.ExtraFiles = append(child.ExtraFiles, []*os.File{childIn, childOut}...)
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	child.Env = append(os.Environ(), []string{
		"__privsep_phase=dropping",
		"__privsep_user=" + username,
	}...)

	if err = child.Run(); err != nil {
		// phase 1 child only writes to childOut when there's an error
		line, _ := bufio.NewReader(parentIn).ReadString('\n')
		err = fmt.Errorf("couldn't drop privileges: %q", line)
		return
	}

	// parent doesn't need these anymore
	childIn.Close()
	childOut.Close()

	r = parentIn
	w = parentOut

	return
}

func maybeBecomeChild() (isChild bool, r io.Reader, w io.Writer, err error) {

	// dropping privileges is a two-phase process since a Go program cannot
	// completely drop privileges after the runtime has started; only the
	// thread which calls setuid will have its uid changed, and there is no way
	// to iterate over all the runtime's threads to make that happen
	// process-wide. instead, the child must exec itself on the same thread
	// that calls setuid; the new runtime's threads will then all be owned by
	// the target user.

	switch os.Getenv("__privsep_phase") {

	default:
		// not the child
		return

	case "dropping":
		// phase 1: we're the child, but haven't dropped privileges

		defer os.Exit(0) // never return to caller from this phase

		var bin string
		bin, err = executable.Path()
		if err != nil {
			reportError(err)
		}

		// make sure the thread that exec's is the same one that drops privs
		runtime.LockOSThread()

		if err = dropPrivs(); err != nil {
			reportError(err)
		}

		cleanEnv()
		os.Setenv("__privsep_phase", "dropped")

		attr := syscall.ProcAttr{
			Env: os.Environ(),
			Files: []uintptr{
				0, 1, 2, // inherit stdin/out/err
				3, 4, // childIn and childOut from createChild
			},
		}

		// ideally we could just exec so the parent could wait() for the final
		// child, but syscall.Exec doesn't accept a ProcAttr, so instead use
		// StartProcess which forks then execs
		args := append([]string{bin}, origArgs[1:]...)
		if _, _, err = syscall.StartProcess(bin, args, &attr); err != nil {
			reportError(err)
		}

	case "dropped":
		// phase 2: we're the child, now without privileges

		isChild = true

		if os.Getuid() == 0 {
			err = errors.New("child is still privileged")
			return
		}

		cleanEnv()

		r = os.NewFile(3, "input")
		w = os.NewFile(4, "output")
	}

	return
}

func dropPrivs() error {
	username := os.Getenv("__privsep_user")

	if username == "" {
		return errors.New("no __privsep_user")
	}

	u, err := user.Lookup(username)
	if err != nil {
		return err
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return fmt.Errorf("uid of %s isn't numeric: %q", u.Uid, u.Uid)
	}

	// XXX can't lookup gid by name, http://code.google.com/p/go/issues/detail?id=2617
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return fmt.Errorf("gid of %s isn't numeric: %q", u.Gid, u.Gid)
	}

	// change gid first since it can't be changed after dropping root uid
	if err := syscall.Setgid(gid); err != nil {
		return err
	}
	if err := syscall.Setuid(uid); err != nil {
		return err
	}

	return nil
}

func reportError(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.NewFile(4, ""), err.Error())
	os.Exit(1)
}
