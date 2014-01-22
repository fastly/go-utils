package executable_test

import (
	"github.com/fastly/go-utils/executable"
	"strings"
	"testing"
)

// tests Path as well
func TestDir(t *testing.T) {
	expectContains := "github.com/fastly/go-utils/executable"
	dir, err := executable.Dir()
	if err != nil {
		t.Fatalf("unable to get test dir, err: %v", err)
	}
	if !strings.Contains(dir, expectContains) {
		t.Errorf("wrong executable dir, got: %v, expectedContains: %v", dir, expectContains)
	}
}
