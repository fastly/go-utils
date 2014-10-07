package privsep

// TODO:
// - Test various failures in child
// - Check for presence and order of arguments as seen in child

import (
	"bufio"
	"io"
	"log"
	"os"
	"testing"
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
		t.Skip("test must run as root: go test -c github.com/fastly/go-utils/privsep && sudo ./privsep.test -test.v")
	}

	// flags should be passed through and correctly ordered
	args := []string{"--flag", "arg"}

	// extra fds should show up in order
	r1, w1, err := os.Pipe()
	if err != nil {
		t.Errorf("Pipe: %s", err)
	}

	pid, r, w, err := CreateChild("nobody", os.Args[0], args, []*os.File{w1})
	if err != nil {
		t.Fatalf("CreateChild failed: %s", err)
	}

	defer func() {
		if p, err := os.FindProcess(pid); err == nil {
			p.Kill()
		}
	}()

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

func child(r io.Reader, w, w1 io.Writer) {
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
