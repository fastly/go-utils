// +build linux darwin

package strftime_test

import (
	"testing"

	"github.com/fastly/go-utils/strftime"
)

// these test the pure Go implementation on platforms where the POSIX version
// is used for strftime.Strftime

func TestStrftimePureAgainstReference(t *testing.T) {
	testStrftimeAgainstReference(t, strftime.StrftimePure, true)
}

func TestStrftimePureAgainstPerl(t *testing.T) {
	testStrftimeAgainstPerl(t, strftime.StrftimePure, true)
}
