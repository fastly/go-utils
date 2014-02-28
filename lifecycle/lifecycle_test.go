package lifecycle

import (
	"os"
	"testing"
	"time"
)

func TestRunWhenKilled(t *testing.T) {
	i := 0
	finalFunc := func() {
		i++
	}
	l := New(true)
	go func() {
		l.RunWhenKilled(finalFunc, 100*time.Millisecond)
	}()
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("%v", err)
	}
	proc.Signal(os.Interrupt)
	time.Sleep(time.Millisecond)
	if i != 1 {
		t.Errorf("got %v != expect 1", i)
	}
}
