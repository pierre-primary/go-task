package lock

import (
	"runtime"
	"sync/atomic"
)

const (
	_LOCK_MASK uint32 = ^uint32(0)
)

type SpinRWLock struct {
	state atomic.Uint32
}

func (l *SpinRWLock) Lock() {
	for !l.state.CompareAndSwap(0, _LOCK_MASK) {
		runtime.Gosched()
	}
}

func (l *SpinRWLock) TryLock() bool {
	return l.state.CompareAndSwap(0, _LOCK_MASK)
}

func (l *SpinRWLock) Unlock() {
	l.state.Store(0)
}

func (l *SpinRWLock) RLock() {
	for {
		state := l.state.Load()
		if state == _LOCK_MASK {
			runtime.Gosched()
		} else if l.state.CompareAndSwap(state, state+1) {
			return
		}
	}
}

func (l *SpinRWLock) TryRLock() bool {
	for {
		state := l.state.Load()
		if state == _LOCK_MASK {
			return false
		} else if l.state.CompareAndSwap(state, state+1) {
			return true
		}
	}
}

func (l *SpinRWLock) RUnlock() {
	l.state.Add(_LOCK_MASK)
}
