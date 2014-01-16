package silence

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type SilencerState struct {
	next     time.Time
	count    int
	lastFunc func(int, string)
}

type Silencer struct {
	sync.Mutex
	states map[string]*SilencerState
}

func NewSilencer() *Silencer {
	return &Silencer{states: make(map[string]*SilencerState)}
}

var silencer = NewSilencer()

// SilenceFor aggregates repeated calls to itself and calls f once every duration
// with the number of aggregated calls and the id tagged with the calling file and
// line number
func SilenceFor(duration time.Duration, id string, f func(int, string)) {
	// WrapSilenceFor is depth 1,
	WrapSilenceFor(2, duration, id, f)
}

// WrapSilenceFor is the same as SilenceFor, except the depth of the call stack can be
// chosen for what ID to tag and collate. runtime.Caller will have depth 0, and the
// call to this function will have depth 1, so any additional layers before calling
// this function should have depth >= 1.
func WrapSilenceFor(depth int, duration time.Duration, id string, f func(int, string)) {
	_, file, line, _ := runtime.Caller(depth)
	file = filepath.Base(file)
	key := fmt.Sprintf("%s:%d %s", file, line, id)

	silencer.Lock()
	state, exists := silencer.states[key]
	if !exists {
		state = new(SilencerState)
	}

	state.count++
	state.lastFunc = f

	if state.count == 1 {
		now := time.Now()
		if state.next.Before(now) {
			// no need to suppress
			state.lastFunc(state.count, key)
			state.next = now.Add(duration)
			state.count = 0
		} else {
			// this is the first time we've suppressed, so schedule the next
			// firing. subsequent passes will only increment the counter.
			go func() {
				silencer.Lock()
				wait := state.next.Sub(time.Now())
				silencer.Unlock()

				time.Sleep(wait)

				silencer.Lock()
				state.lastFunc(state.count, key)
				state.next = time.Now().Add(duration)
				state.count = 0
				silencer.states[key] = state
				silencer.Unlock()
			}()
		}
	} // else state.count > 1 and a flush has already been scheduled

	silencer.states[key] = state
	silencer.Unlock()
}
