package task

import (
	"context"
	"time"
)

// ------------------------------------------------------------------------------------------------------
// 实现 Task 接口的方法

func (task *TaskImpl) State() TaskState {
	switch state(task) & maskState {
	case flagCompleted:
		return STATE_COMPLETED
	case flagFailed:
		return STATE_FAULTED
	case flagCanceled:
		return STATE_CANCELED
	default:
		return STATE_PENDING
	}
}

func (task *TaskImpl) IsDone() bool {
	return stateIs(task, checkDone)
}

func (task *TaskImpl) IsCompleted() bool {
	return stateIs(task, checkCompleted)
}

func (task *TaskImpl) IsFaulted() bool {
	return stateIs(task, flagFailed)
}

func (task *TaskImpl) IsCanceled() bool {
	return stateIs(task, flagCanceled)
}

func (task *TaskImpl) Return() (interface{}, error) {
	wait(task)
	if stateIs(task, checkCompleted) {
		return task.data, nil
	}
	err, _ := task.data.(error)
	return nil, err
}

func (task *TaskImpl) Result() interface{} {
	wait(task)
	if stateIs(task, checkCompleted) {
		return task.data
	}
	return nil
}

func (task *TaskImpl) Error() error {
	wait(task)
	if stateIs(task, checkCompleted) {
		return nil
	}
	err, _ := task.data.(error)
	return err
}

func (task *TaskImpl) Done() chan struct{} {
	return done(task)
}

func (task *TaskImpl) Wait(ctxs ...context.Context) Task {
	if task.IsDone() {
		return task
	}
	ctx := firstContext(ctxs, context.TODO())
	select {
	case <-done(task):
	case <-ctx.Done():
	}
	return task
}

func (task *TaskImpl) WaitTimeout(d time.Duration, ctxs ...context.Context) Task {
	if task.IsDone() {
		return task
	}
	ctx := firstContext(ctxs, context.TODO())
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-done(task):
	case <-ctx.Done():
	case <-timer.C:
	}
	return task
}

/* Continue */

type continueCaller ContinueFunc

func (body continueCaller) TryCall(target Task) (bool, interface{}, error) {
	if body == nil {
		return false, nil, nil
	}
	rs, err := body(target)
	return true, rs, err
}

func (task *TaskImpl) Continue(fn ContinueFunc, ctxs ...context.Context) Task {
	ctx := firstContext(ctxs, nil)
	if ctx != nil && isCanceledContext(ctx) {
		return Cancel(ctx.Err())
	}
	return newAsyncFollower(task, continueCaller(fn), ctx)
}

func (task *TaskImpl) ContinueAwait(fn ContinueFunc, ctxs ...context.Context) Task {
	ctx := firstContext(ctxs, nil)
	if ctx != nil && isCanceledContext(ctx) {
		return Cancel(ctx.Err())
	}
	return newSyncFollower(task, continueCaller(fn), ctx)
}

/* Then */

type thenCaller ThenFunc

func (body thenCaller) TryCall(target Task) (bool, interface{}, error) {
	if body == nil || !target.IsCompleted() {
		return false, nil, nil
	}
	rs, err := body(target.Result())
	return true, rs, err
}

func (task *TaskImpl) Then(fn ThenFunc, ctxs ...context.Context) Task {
	ctx := firstContext(ctxs, nil)
	if ctx != nil && isCanceledContext(ctx) {
		return Cancel(ctx.Err())
	}
	return newAsyncFollower(task, thenCaller(fn), ctx)
}

func (task *TaskImpl) ThenAwait(fn ThenFunc, ctxs ...context.Context) Task {
	ctx := firstContext(ctxs, nil)
	if ctx != nil && isCanceledContext(ctx) {
		return Cancel(ctx.Err())
	}
	return newSyncFollower(task, thenCaller(fn), ctx)
}

/* Catch */

type catchCaller CatchFunc

func (body catchCaller) TryCall(target Task) (bool, interface{}, error) {
	if body == nil || !target.IsFaulted() {
		return false, nil, nil
	}
	rs, err := body(target.Error())
	return true, rs, err
}

func (task *TaskImpl) Catch(fn CatchFunc, ctxs ...context.Context) Task {
	ctx := firstContext(ctxs, nil)
	if ctx != nil && isCanceledContext(ctx) {
		return Cancel(ctx.Err())
	}
	return newAsyncFollower(task, catchCaller(fn), ctx)
}

func (task *TaskImpl) CatchAwait(fn CatchFunc, ctxs ...context.Context) Task {
	ctx := firstContext(ctxs, nil)
	if ctx != nil && isCanceledContext(ctx) {
		return Cancel(ctx.Err())
	}
	return newSyncFollower(task, catchCaller(fn), ctx)
}
