// Package suppress has functions to suppress repeated function calls
// into one aggregated function call.
package suppress

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// hide the variable and the struct because they gon' find you
// also because it's meant to be a singleton
var _suppressor *suppressor

type suppressorState struct {
	next     time.Time
	count    int
	lastFunc func(int, string)
}

type suppressor struct {
	sync.Mutex
	states map[string]*suppressorState
}

func init() {
	_suppressor = &suppressor{states: make(map[string]*suppressorState)}
}

// For aggregates repeated calls to itself and calls f once every duration
// with the number of aggregated calls and the tag coalesced with the calling
// file and line number
func For(duration time.Duration, id string, f func(int, string)) {
	// WrapFor is depth 1,
	WrapFor(2, duration, id, f)
}

// WrapFor is the same as For, except the depth of the call stack can be
// chosen for what ID to tag and coalesce. runtime.Caller will have depth 0, and the
// call to this function will have depth 1, so any additional layers before calling
// this function should have depth >= 1.
func WrapFor(depth int, duration time.Duration, id string, f func(int, string)) {
	pc, file, line, _ := runtime.Caller(depth)
	file = filepath.Base(file)
	key := fmt.Sprintf("%d%s", pc, id)
	tag := fmt.Sprintf("%s:%d %s", file, line, id)

	_suppressor.Lock()
	state, exists := _suppressor.states[key]
	if !exists {
		state = new(suppressorState)
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
				_suppressor.Lock()
				wait := state.next.Sub(time.Now())
				_suppressor.Unlock()

				time.Sleep(wait)

				_suppressor.Lock()
				state.lastFunc(state.count, tag)
				state.next = time.Now().Add(duration)
				state.count = 0
				_suppressor.states[key] = state
				_suppressor.Unlock()
			}()
		}
	} // else state.count > 1 and a flush has already been scheduled

	_suppressor.states[key] = state
	_suppressor.Unlock()
}
