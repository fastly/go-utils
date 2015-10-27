package privsep

// TODO:
// - Test various failures in child
// - Check for presence and order of arguments as seen in child

import (
	"bufio"
	"io"
	"log"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"syscall"
	"testing"
	"time"
)

var isChild bool

const (
	ping = "ping\n"
	pong = "pong\n"
	foo  = "foo\n"
)

func TestPrivsep(t *testing.T) {
	if isChild {
		return
	}

	if os.Getuid() != 0 {
		// `sudo go test` doesn't work because it writes the test binary to
		// /tmp as root with 700 permissions
		t.Skip("test must run as root: go test -c github.com/fastly/go-utils/privsep && sudo ./privsep.test -test.v")
	}

	// extra fds should show up in order
	r1, w1, err := os.Pipe()
	if err != nil {
		t.Errorf("Pipe: %s", err)
	}

	proc, r, w, err := CreateChild("nobody", os.Args[0], testArgs, []*os.File{w1})
	if err != nil {
		t.Fatalf("CreateChild failed: %s", err)
	}

	io.WriteString(w, ping)

	for _, e := range envVars {
		if v := os.Getenv(e); v != "" {
			t.Errorf("%s env var should be empty, is %q", e, v)
		}
	}

	// check default pipe
	br := bufio.NewReader(r)
	reply, _ := br.ReadString('\n')
	if reply == pong {
		t.Logf("got expected reply %q", reply)
	} else {
		t.Errorf("expected %q, got %q", pong, reply)
	}

	// check extra pipe
	br = bufio.NewReader(r1)
	reply, _ = br.ReadString('\n')
	if reply == foo {
		t.Logf("got expected reply %q", reply)
	} else {
		t.Errorf("expected %q, got %q", foo, reply)
	}

	c := make(chan *os.ProcessState, 1)
	go func() {
		proc.Kill()
		state, _ := proc.Wait()
		c <- state
	}()

	select {
	case state := <-c:
		if status, ok := state.Sys().(syscall.WaitStatus); ok {
			t.Logf("child exited with status %d", status.ExitStatus())
		} else {
			t.Log("child exited")
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for child to die")
	}

	maxFd, err := getHighestFd()
	if err != nil {
		t.Error(err)
	}

	// expect stdin, stdout, stderr, r1, w1, r, w
	if maxFd != 7 {
		t.Errorf("wanted maximum fd of %d, got %d", 7, maxFd)
	}

	r.(*os.File).Close()
	w.(*os.File).Close()
	r1.Close()
	w1.Close()

	if maxFd, err = getHighestFd(); err != nil {
		t.Error(err)
	}

	// now just stdin, stdout, stderr
	if maxFd != 3 {
		t.Errorf("wanted maximum fd of %d, got %d", 3, maxFd)
	}
}

func init() {
	is, r, w, files, err := MaybeBecomeChild()
	isChild = is
	if err != nil {
		log.Fatalf("MaybeBecomeChild: %s", err)
	}
	if isChild {
		child(r, w, files[0])
	}
}

var testArgs = []string{"--flag", "arg"}

func child(r io.Reader, w, w1 io.Writer) {
	args := os.Args[1:]
	if !reflect.DeepEqual(args, testArgs) {
		log.Fatalf("got args %+v, expected %+v", args, testArgs)
	}

	for _, e := range envVars {
		if v := os.Getenv(e); v != "" {
			log.Fatalf("%s env var should be empty, is %q", e, v)
		}
	}

	br := bufio.NewReader(r)
	line, _ := br.ReadString('\n')
	if line != ping {
		log.Fatalf("expected %q, got %q", ping, line)
	}

	maxFd, err := getHighestFd()
	if err != nil {
		log.Fatal(err)
	}

	// expect stdin, stdout, stderr, r, w, w1
	if maxFd != 6 {
		log.Fatalf("wanted maximum fd of %d, got %d", 6, maxFd)
	}

	io.WriteString(w, pong)
	io.WriteString(w1, foo)

	os.Exit(0)
}

// getHighestFd returns the highest valued file descriptor open in the current
// process
func getHighestFd() (int, error) {
	var fdPath string
	switch runtime.GOOS {
	case "linux":
		fdPath = "/proc/self/fd/"
	case "darwin":
		fdPath = "/dev/fd/"
	default:
		panic("unsupported platform")
	}
	dh, err := os.Open(fdPath)
	if err != nil {
		return -1, err
	}
	defer dh.Close()
	names, err := dh.Readdirnames(0)
	if err != nil {
		return -1, err
	}
	var max int
	for _, name := range names {
		if n, err := strconv.Atoi(name); err == nil && n > max {
			max = n
		}
	}
	return max, nil
}
