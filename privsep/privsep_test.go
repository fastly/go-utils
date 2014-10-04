package privsep

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
)

func TestPrivsep(t *testing.T) {
	if isChild {
		return
	}

	if os.Getuid() != 0 {
		t.Skip("test must run as root: go test -c github.com/fastly/go-utils/privsep && sudo ./privsep.test -test.v")
	}

	r, w, err := CreateChild("nobody", os.Args[0])
	if err != nil {
		t.Fatalf("CreateChild failed: %s", err)
	}
	io.WriteString(w, ping)

	for _, e := range []string{"__privsep_phase", "__privsep_user"} {
		if v := os.Getenv(e); v != "" {
			t.Errorf("%s env var should be empty, is %q", e, v)
		}
	}

	br := bufio.NewReader(r)
	reply, _ := br.ReadString('\n')
	if reply == pong {
		t.Logf("got expected reply %q", reply)
	} else {
		t.Errorf("expected %q, got %q", pong, reply)
	}
}

func init() {
	is, r, w, err := MaybeBecomeChild()
	isChild = is
	if err != nil {
		log.Fatalf("MaybeBecomeChild: %s", err)
	}
	if isChild {
		child(r, w)
	}
}

func child(r io.Reader, w io.Writer) {
	for _, e := range []string{"__privsep_phase", "__privsep_user"} {
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
	os.Exit(0)
}
