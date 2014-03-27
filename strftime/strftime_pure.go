package strftime

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// StrftimePure is a locale-unaware implementation of strftime(3). It does not
// correctly account for locale-specific conversion specifications, so formats
// like `%c` may behave differently from the underlying platform. Additionally,
// the `%E` and `%O` modifiers are passed through as raw strings.
//
// The implementation of locale-specific conversions attempts to mirror the
// strftime(3) implementation in glibc 2.15 under LC_TIME=C.
func StrftimePure(format string, t time.Time) string {
	elts := make([]string, 0)
	for i := 0; i < len(format); {
		var e string
		if format[i] == '%' {
			i++
			if i < len(format) {
				switch format[i] {
				default:
					e = format[i-1:i+1]
				case 'a':
					// The abbreviated weekday name according to the current locale.
					e = t.Format("Mon")
				case 'A':
					// The full weekday name according to the current locale.
					e = t.Format("Monday")
				case 'b':
					// The abbreviated month name according to the current locale.
					e = t.Format("Jan")
				case 'B':
					// The full month name according to the current locale.
					e = t.Month().String()
				case 'C':
					// The century number (year/100) as a 2-digit integer. (SU)
					e = strconv.Itoa(int(t.Year() / 100))
				case 'c':
					// The preferred date and time representation for the current locale.
					e = t.Format("Mon Jan  2 15:04:05 2006")
				case 'd':
					// The day of the month as a decimal number (range 01 to 31).
					e = fmt.Sprintf("%02d", t.Day())
				case 'D':
					// Equivalent to %m/%d/%y. (Yecch—for Americans only. Americans should note that in other countries %d/%m/%y is rather com‐ mon. This means that in international context this format is ambiguous and should not be used.) (SU)
					e = fmt.Sprintf("%02d/%02d/%02d", t.Month(), t.Day(), t.Year()%100)
				case 'E':
					// Modifier: use alternative format, see below. (SU)
					if i+1 < len(format) {
						i++
						e = fmt.Sprintf("%%E%c", format[i])
					} else {
						e = "%E"
					}
				case 'e':
					// Like %d, the day of the month as a decimal number, but a leading zero is replaced by a space. (SU)
					e = fmt.Sprintf("%2d", t.Day())
				case 'F':
					// Equivalent to %Y-%m-%d (the ISO 8601 date format). (C99)
					e = fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
				case 'G':
					// The ISO 8601 week-based year (see NOTES) with century as a decimal number. The 4-digit year corresponding to the ISO week number (see %V). This has the same format and value as %Y, except that if the ISO week number belongs to the previous or next year, that year is used instead. (TZ)
					year, _ := t.ISOWeek()
					e = fmt.Sprintf("%04d", year)
				case 'g':
					// Like %G, but without century, that is, with a 2-digit year (00-99). (TZ)
					year, _ := t.ISOWeek()
					e = fmt.Sprintf("%02d", year%100)
				case 'h':
					// Equivalent to %b. (SU)
					e = t.Format("Jan")
				case 'H':
					// The hour as a decimal number using a 24-hour clock (range 00 to 23).
					//e = fmt.Sprintf("%02d", t.Hour())
					e = fmt.Sprintf("%02d", t.Hour())
				case 'I':
					// The hour as a decimal number using a 12-hour clock (range 01 to 12).
					e = fmt.Sprintf("%02d", t.Hour()%12)
				case 'j':
					// The day of the year as a decimal number (range 001 to 366).
					e = fmt.Sprintf("%03d", t.YearDay())
				case 'k':
					// The hour (24-hour clock) as a decimal number (range 0 to 23); single digits are preceded by a blank. (See also %H.) (TZ)
					e = fmt.Sprintf("%2d", t.Hour())
				case 'l':
					// The hour (12-hour clock) as a decimal number (range 1 to 12); single digits are preceded by a blank. (See also %I.) (TZ)
					e = fmt.Sprintf("%2d", t.Hour()%12)
				case 'm':
					// The month as a decimal number (range 01 to 12).
					e = t.Format("01")
				case 'M':
					// The minute as a decimal number (range 00 to 59).
					e = t.Format("04")
				case 'n':
					// A newline character. (SU)
					e = "\n"
				case 'O':
					// Modifier: use alternative format, see below. (SU)
					if i+1 < len(format) {
						i++
						e = fmt.Sprintf("%%O%c", format[i])
					} else {
						e = "%O"
					}
				case 'p':
					// Either "AM" or "PM" according to the given time value, or the corresponding strings for the current locale. Noon is treated as "PM" and midnight as "AM".
					e = t.Format("PM")
				case 'P':
					// Like %p but in lowercase: "am" or "pm" or a corresponding string for the current locale. (GNU)
					e = t.Format("pm")
				case 'r':
					// The time in a.m. or p.m. notation. In the POSIX locale this is equivalent to %I:%M:%S %p. (SU)
					e = t.Format("03:04:05 PM")
				case 'R':
					// The time in 24-hour notation (%H:%M). (SU) For a version including the seconds, see %T below.
					e = t.Format("15:04")
				case 's':
					// The number of seconds since the Epoch, 1970-01-01 00:00:00 +0000 (UTC). (TZ)
					e = fmt.Sprintf("%d", t.Unix())
				case 'S':
					// The second as a decimal number (range 00 to 60). (The range is up to 60 to allow for occasional leap seconds.)
					e = t.Format("05")
				case 't':
					// A tab character. (SU)
					e = "\t"
				case 'T':
					// The time in 24-hour notation (%H:%M:%S). (SU)
					e = t.Format("15:04:05")
				case 'u':
					// The day of the week as a decimal, range 1 to 7, Monday being 1. See also %w. (SU)
					day := int(t.Weekday())
					if day == 0 {
						day = 7
					}
					e = strconv.Itoa(day)
				case 'U':
					// The week number of the current year as a decimal number, range 00 to 53, starting with the first Sunday as the first day of week 01. See also %V and %W.
					e = fmt.Sprintf("%02d", (t.YearDay() - int(t.Weekday()) + 7) / 7)
				case 'V':
					// The ISO 8601 week number (see NOTES) of the current year as a decimal number, range 01 to 53, where week 1 is the first week that has at least 4 days in the new year. See also %U and %W. (SU)
					_, week := t.ISOWeek()
					e = fmt.Sprintf("%02d", week)
				case 'w':
					// The day of the week as a decimal, range 0 to 6, Sunday being 0. See also %u.
					e = strconv.Itoa(int(t.Weekday()))
				case 'W':
					// The week number of the current year as a decimal number, range 00 to 53, starting with the first Monday as the first day of week 01.
					e = fmt.Sprintf("%02d", (t.YearDay() - (int(t.Weekday()) - 1 + 7) % 7 + 7) / 7)
				case 'x':
					// The preferred date representation for the current locale without the time.
					e = fmt.Sprintf("%02d/%02d/%02d", t.Month(), t.Day(), t.Year()%100)
				case 'X':
					// The preferred time representation for the current locale without the date.
					e = t.Format("15:04:05")
				case 'y':
					// The year as a decimal number without a century (range 00 to 99).
					e = fmt.Sprintf("%02d", t.Year()%100)
				case 'Y':
					// The year as a decimal number including the century.
					e = t.Format("2006")
				case 'z':
					// The +hhmm or -hhmm numeric timezone (that is, the hour and minute offset from UTC). (SU)
					e = t.Format("-0700")
				case 'Z':
					// The timezone or name or abbreviation.
					e = t.Format("MST")
				case '+':
					// The date and time in date(1) format. (TZ) (Not supported in glibc2.)
					e = t.Format("Mon Jan _2 15:04:05 MST 2006")
				case '%':
					// A literal '%' character.
					e = "%"
				}
				i++
			} else {
				// pass through a % at the end of the format string
				e = "%"
			}
		} else {
			nextPct := strings.Index(format[i:], "%")
			if nextPct >= 0 {
				e = format[i:i+nextPct]
				i += len(e)
			} else {
				e = format[i:]
				i = len(format)
			}
		}
		elts = append(elts, e)
	}
	return strings.Join(elts, "")
}
