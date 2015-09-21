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

	io.WriteString(w, pong)
	io.WriteString(w1, foo)

	os.Exit(0)
}
