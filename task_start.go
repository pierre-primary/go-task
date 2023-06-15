package task

import (
	"context"
	"sync"
)

type (
	Starter struct {
		task   *TaskImpl
		ctx    context.Context
		caller StartCaller
	}

	StartCaller interface {
		Call(task *TaskImpl)
	}
)

var starterPool = sync.Pool{
	New: func() interface{} {
		return new(Starter)
	},
}

// Release starter object
func (starter *Starter) release() {
	starter.task = nil
	starter.caller = nil
	starterPool.Put(starter)
}

// Assign a starter object
func assignStarter(task *TaskImpl, caller StartCaller, ctx context.Context) *Starter {
	starter := starterPool.Get().(*Starter)
	starter.task = task
	starter.caller = caller
	starter.ctx = ctx
	return starter
}

// Asynchronous execute starter task
func asyncExecStarter(starter *Starter) {
	go func() {
		// small closure way
		// donot modify closure variable (starter),
		// avoid additional memory allocation.
		task := starter.task
		caller := starter.caller
		ctx := starter.ctx
		starter.release()

		// async check context is canceled
		if ctx != nil && isCanceledContext(ctx) {
			cancel(task, ctx.Err())
			return
		}

		syncExecStarter(task, caller)
	}()
}

// Synchronous execute starter task
func syncExecStarter(task *TaskImpl, caller StartCaller) {
	// safe exit
	defer func() {
		if r := recover(); r != nil {
			reject(task, r)
		}
	}()
	// call
	caller.Call(task)
}

// Create a starter task
func newStarter(caller StartCaller, ctx context.Context) *TaskImpl {
	task := newTask()
	starter := assignStarter(task, caller, ctx)
	// async exec
	asyncExecStarter(starter)
	return task
}
