package silence_test

import (
	"github.com/fastly/go-utils/silence"
	"runtime"
	"testing"
	"time"
)

/*
func TestSilencer1(t *testing.T) {
	test(t, []string{""}, 1, 1000*time.Millisecond, 100*time.Millisecond, 11)
}

func TestSilencer2(t *testing.T) {
	test(t, []string{"spot1", "spot2"}, 1, 1000*time.Millisecond, 100*time.Millisecond, 11)
}

func TestSilencer3(t *testing.T) {
	test(t, []string{""}, 3, 1000*time.Millisecond, 100*time.Millisecond, 11)
}

func TestSilencer4(t *testing.T) {
	test(t, []string{"#1", "#2"}, 3, 1000*time.Millisecond, 100*time.Millisecond, 11)
}

func TestSilencer5(t *testing.T) {
	test(t, []string{"#1", "#2"}, 3, 10*time.Millisecond, 100*time.Millisecond, 2)
}
*/

func test(t *testing.T, tags []string, invocations int, testTime time.Duration, silenceTime time.Duration, expectedPerInvocation int) {
	var attempts, firings, errors int
	lasts := make(map[string]time.Time)

	f := func(count int, id string) {
		last := lasts[id]
		now := time.Now()
		if last.IsZero() || now.Sub(last) >= silenceTime {
			firings++
		} else {
			t.Logf("Error %q at %v; delta=%v attempts=%d last=%v count=%d",
				id, time.Now(), now.Sub(last), attempts, last, count)
			errors++
		}
		lasts[id] = now
	}

	expected := invocations * len(tags) * expectedPerInvocation

	start := time.Now()
	end := start.Add(testTime * 11 / 10)
	for time.Now().Before(end) {
		attempts++
		if firings >= expected {
			break
		}
		tag := tags[attempts%len(tags)]
		// use separate calls so program counter is different for each
		if invocations > 0 {
			SilenceFor(silenceTime, tag, f)
		}
		if invocations > 1 {
			SilenceFor(silenceTime, tag, f)
		}
		if invocations > 2 {
			SilenceFor(silenceTime, tag, f)
		}
		runtime.Gosched() // yield to other goroutines
	}

	// wait for flusher goroutines to finish
	finished := time.Now()
	time.Sleep(2 * silenceTime)
	finishedAndWaited := time.Now()

	elapsed := finished.Sub(start)
	longElapsed := finishedAndWaited.Sub(start)

	t.Logf("Ran %d iterations in %v (%v with wait), fired correctly %d times (wanted %d) and %d incorrectly",
		attempts, elapsed, longElapsed, firings, expected, errors)
	if firings != expected {
		t.Errorf("Expected %d firings, got %d", expected, firings)
	}
	if errors > 0 {
		t.Errorf("Silencer failed to silence %d times", errors)
	}
}

func TestSilencerStalled(t *testing.T) {
	type Event struct {
		time time.Time
		n    int
	}
	events := make([]Event, 0)

	// fire 5 events in rapid succession, all within the silence window. the
	// first call should happen immediately but the next four should be
	// coalesced at the end of the silence period.
	start := time.Now()
	for i := 0; i < 5; i++ {
		SilenceFor(100*time.Millisecond, "anon", func(n int, tag string) {
			events = append(events, Event{time.Now(), n})
			t.Logf("%v", tag)
		})
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(start.Add(110 * time.Millisecond).Sub(time.Now()))

	if len(events) != 2 || events[0].n != 1 || events[1].n != 4 {
		t.Errorf("unexpected event stream: %+v", events)
	}
}
