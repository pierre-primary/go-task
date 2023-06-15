package task

import (
	"context"
	"sync"
)

type (
	Follower struct {
		task   *TaskImpl
		ctx    context.Context
		caller FollowCaller
		next   *Follower
	}

	FollowCaller interface {
		TryCall(target Task) (bool, interface{}, error)
	}
)

var flwPool = sync.Pool{
	New: func() interface{} {
		return new(Follower)
	},
}

// Release follower object
func (flw *Follower) release() {
	flw.task = nil
	flw.caller = nil
	flw.ctx = nil
	flw.next = nil
	flwPool.Put(flw)
}

// Assign a follower object
func assignFollower(task *TaskImpl, caller FollowCaller, ctx context.Context) *Follower {
	flw := flwPool.Get().(*Follower)
	flw.task = task
	flw.caller = caller
	flw.ctx = ctx
	return flw
}

// Wake up all follower
func wakeAllFollower(target *TaskImpl) {
	flw := target.flws
	for flw != nil {
		next := flw.next
		flw.next = nil
		asyncExecFollower(flw, target, true)
		flw = next
	}
}

// Asynchronous execute follower task
func asyncExecFollower(flw *Follower, target Task, syncCheckConetxt bool) {
	// sync check context is canceled
	if syncCheckConetxt && flw.ctx != nil && isCanceledContext(flw.ctx) {
		cancel(flw.task, flw.ctx.Err())
		return
	}
	go func() {
		// small closure way
		// donot modify closure variable (flw, target),
		// avoid additional memory allocation.
		task := flw.task
		caller := flw.caller
		ctx := flw.ctx
		flw.release()

		// async check context is canceled
		if ctx != nil && isCanceledContext(ctx) {
			cancel(task, ctx.Err())
			return
		}

		syncExecFollower(task, caller, target)
	}()
}

// Synchronous execute follower task
func syncExecFollower(task *TaskImpl, caller FollowCaller, target Task) {
	// safe exit
	done := false
	defer func() {
		if r := recover(); r != nil {
			reject(task, r)
		} else if !done {
			resolve(task, nil)
		}
	}()
	// try call
	if ok, rs, err := caller.TryCall(target); ok {
		// can handle
		if err == nil {
			resolve(task, rs)
		} else {
			reject(task, err)
		}
	} else {
		// canot handle, passing the result
		switch target.State() {
		case STATE_COMPLETED:
			resolve(task, target.Result())
		case STATE_CANCELED:
			cancel(task, target.Error())
		default:
			reject(task, target.Error())
		}
	}
	done = true
}

// Create a sync follower task, sync wait and execute
func newSyncFollower(task *TaskImpl, caller FollowCaller, ctx context.Context) *TaskImpl {
	flwTask := newTask()
	if ctx == nil {
		ctx = context.TODO()
	}
	select {
	case <-ctx.Done():
		cancel(flwTask, ctx.Err())
		return flwTask
	case <-done(task):
	}
	syncExecFollower(flwTask, caller, task)
	return flwTask
}

// Create a async follower task
func newAsyncFollower(task *TaskImpl, caller FollowCaller, ctx context.Context) *TaskImpl {
	flwTask := newTask()
	flw := assignFollower(flwTask, caller, ctx)
	// try join in follower linked
	if lockStateIfNot(task, lockFlws, checkDone) {
		flw.next = task.flws
		task.flws = flw
		unlockStateAndSet(task, lockFlws, 0)
		return flwTask
	}
	// join in failed，async exec
	asyncExecFollower(flw, task, false)
	return flwTask
}
