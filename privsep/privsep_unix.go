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
	"strings"
	"syscall"

	"github.com/fastly/go-utils/executable"
)

const (
	exitOK               = 0
	exitOSFailure        = 1
	exitDescribedFailure = 2
	exitFailsafe         = 3
)

func createChild(username, bin string, args []string, files []*os.File) (pid int, r io.Reader, w io.Writer, err error) {
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

	child := exec.Command(bin, args...)

	// childIn becomes fd 3 in child, childOut becomes fd 4, etc
	child.ExtraFiles = append(child.ExtraFiles, []*os.File{childIn, childOut}...)
	if len(files) > 0 {
		child.ExtraFiles = append(child.ExtraFiles, files...)
	}
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	child.Env = append(os.Environ(), []string{
		"__privsep_phase=dropping",
		"__privsep_user=" + username,
		"__privsep_fds=" + strconv.Itoa(len(files)),
	}...)

	err = child.Start()
	if err != nil {
		err = fmt.Errorf("couldn't start child: %s", err)
		return
	}

	var childReply string

	err = child.Wait()
	if err != nil {
		status := child.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
		if status == exitDescribedFailure {
			if childReply, err = bufio.NewReader(parentIn).ReadString('\n'); err == nil {
				err = fmt.Errorf("failure starting unprivileged child: %s", strings.TrimSpace(childReply))
				return
			}
		}
		err = fmt.Errorf("unknown error starting unprivileged child: %s", err)
	}

	if childReply, err = bufio.NewReader(parentIn).ReadString('\n'); err == nil {
		pid, err = strconv.Atoi(strings.TrimSpace(childReply))
		if err != nil {
			err = fmt.Errorf("bad response from child: %s", err)
			return
		}
	}

	// parent doesn't need these anymore
	childIn.Close()
	childOut.Close()

	r = parentIn
	w = parentOut

	return
}

func maybeBecomeChild() (isChild bool, r io.Reader, w io.Writer, files []*os.File, err error) {

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

		defer os.Exit(exitFailsafe) // never return to caller from this phase

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

		fds := []uintptr{
			0, 1, 2, // inherit stdin/out/err
			3, 4, // childIn and childOut from createChild
			// user fds start at 5
		}
		sfds := os.Getenv("__privsep_fds")
		nfds, _ := strconv.Atoi(sfds)
		for i := 0; i < nfds; i++ {
			fds = append(fds, uintptr(i+5))
		}

		cleanEnv()
		os.Setenv("__privsep_phase", "dropped")
		os.Setenv("__privsep_fds", sfds)

		attr := syscall.ProcAttr{Env: os.Environ(), Files: fds}

		// ideally we could just exec so the parent could wait() for the final
		// child, but syscall.Exec doesn't accept a ProcAttr, so instead use
		// StartProcess which forks then execs
		args := append([]string{bin}, origArgs[1:]...)
		var pid int
		if pid, _, err = syscall.StartProcess(bin, args, &attr); err != nil {
			reportError(err)
		}

		replyToParent(strconv.Itoa(pid))

	case "dropped":
		// phase 2: we're the child, now without privileges

		isChild = true

		if os.Getuid() == 0 {
			err = errors.New("child is still privileged")
			return
		}

		nfds, _ := strconv.Atoi(os.Getenv("__privsep_fds"))

		cleanEnv()

		r = os.NewFile(3, "input")
		w = os.NewFile(4, "output")

		if nfds > 0 {
			files = make([]*os.File, nfds)
			for i := 0; i < nfds; i++ {
				files[i] = os.NewFile(uintptr(i)+5, "output")
			}
		}
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
	replyToParent(err.Error())
	os.Exit(exitDescribedFailure)
}

func replyToParent(reply string) {
	fmt.Fprintln(os.NewFile(4, ""), reply)
}
