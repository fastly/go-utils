// Package silence has functions to suppress repeated function calls
// into one aggregated function call.
package silence

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// hide the variable and the struct because they gon' find you
// also because it's meant to be a singleton
var _silencer *silencer

type silencerState struct {
	next     time.Time
	count    int
	lastFunc func(int, string)
}

type silencer struct {
	sync.Mutex
	states map[string]*silencerState
}

func init() {
	_silencer = &silencer{states: make(map[string]*silencerState)}
}

// For aggregates repeated calls to itself and calls f once every duration
// with the number of aggregated calls and the tag collated with the calling
// file and line number
func For(duration time.Duration, id string, f func(int, string)) {
	// WrapFor is depth 1,
	WrapFor(2, duration, id, f)
}

// WrapFor is the same as For, except the depth of the call stack can be
// chosen for what ID to tag and collate. runtime.Caller will have depth 0, and the
// call to this function will have depth 1, so any additional layers before calling
// this function should have depth >= 1.
func WrapFor(depth int, duration time.Duration, id string, f func(int, string)) {
	pc, file, line, _ := runtime.Caller(depth)
	file = filepath.Base(file)
	key := fmt.Sprintf("%d%s", pc, id)
	tag := fmt.Sprintf("%s:%d %s", file, line, id)

	_silencer.Lock()
	state, exists := _silencer.states[key]
	if !exists {
		state = new(silencerState)
	}

	state.count++
	state.lastFunc = f

	if state.count == 1 {
		now := time.Now()
		if state.next.Before(now) {
			// no need to suppress
			state.lastFunc(state.count, tag)
			state.next = now.Add(duration)
			state.count = 0
		} else {
			// this is the first time we've suppressed, so schedule the next
			// firing. subsequent passes will only increment the counter.
			go func() {
				_silencer.Lock()
				wait := state.next.Sub(time.Now())
				_silencer.Unlock()

				time.Sleep(wait)

				_silencer.Lock()
				state.lastFunc(state.count, tag)
				state.next = time.Now().Add(duration)
				state.count = 0
				_silencer.states[key] = state
				_silencer.Unlock()
			}()
		}
	} // else state.count > 1 and a flush has already been scheduled

	_silencer.states[key] = state
	_silencer.Unlock()
}
