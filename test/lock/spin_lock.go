package lock

import (
	"runtime"
	"sync/atomic"
)

type SpinLock struct {
	state atomic.Uint32
}

func (l *SpinLock) Lock() {
	for !l.state.CompareAndSwap(0, 1) {
		runtime.Gosched()
	}
}

func (l *SpinLock) TryLock() bool {
	return l.state.CompareAndSwap(0, 1)
}

func (l *SpinLock) Unlock() {
	l.state.Store(0)
}
