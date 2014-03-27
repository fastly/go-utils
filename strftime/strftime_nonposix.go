// +build !linux,!darwin

package strftime

import "time"

func strftime(format string, t time.Time) string {
	return StrftimePure(format, t)
}
