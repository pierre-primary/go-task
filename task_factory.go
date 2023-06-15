package task

import (
	"context"
	"time"
)

var (
	defaultResolveTask  = newDoneTask(flagCompleted, nil)
	defaultRejectTask   = newDoneTask(flagFailed, nil)
	defaultCanceledTask = newDoneTask(flagCanceled, canceledError)
	contextCanceledTask = newDoneTask(flagCanceled, context.Canceled)
	contextTimeoutTask  = newDoneTask(flagCanceled, context.DeadlineExceeded)
	nilCanceledTask     = newDoneTask(flagCanceled, nil)
)

func Resolve(result ...interface{}) Task {
	if len(result) == 0 || result[0] == nil {
		return defaultResolveTask
	}
	task := newTask()
	resolve(task, result[0])
	return task
}

func Reject(err ...error) Task {
	if len(err) == 0 || err[0] == nil {
		return defaultRejectTask
	}
	task := newTask()
	reject(task, err[0])
	return task
}

func Cancel(cause ...error) Task {
	if len(cause) == 0 {
		return defaultCanceledTask
	}
	switch {
	case cause[0] == nil:
		return nilCanceledTask
	case cause[0] == context.Canceled:
		return contextCanceledTask
	case cause[0] == context.DeadlineExceeded:
		return contextTimeoutTask
	}
	task := newTask()
	cancel(task, cause[0])
	return task
}

// ------------------------------------------------------------------------------------------
type (
	ResolveFunc = func(interface{})
	RejectFunc  = func(error)
	TaskFunc    = func(*TaskImpl)
)

type taskCaller TaskFunc

func (body taskCaller) Call(task *TaskImpl) {
	body(task)
}

/* New */
func New() (Task, ResolveFunc, RejectFunc) {
	task := newTask()
	return task, func(result interface{}) {
			resolve(task, result)
		}, func(err error) {
			reject(task, err)
		}
}

/* Start */
type StartFunc func(ResolveFunc, RejectFunc)
type startCaller StartFunc

func (body startCaller) Call(task *TaskImpl) {
	body(func(result interface{}) {
		resolve(task, result)
	}, func(err error) {
		reject(task, err)
	})
}

func Start(fn StartFunc, ctxs ...context.Context) Task {
	ctx := firstContext(ctxs, nil)
	if ctx != nil && isCanceledContext(ctx) {
		return Cancel(ctx.Err())
	}
	return newStarter(startCaller(fn), ctx)
}

/* Run */
type RunFunc = func() (interface{}, error)
type runCaller RunFunc

func (body runCaller) Call(task *TaskImpl) {
	if rs, err := body(); err == nil {
		resolve(task, rs)
	} else {
		reject(task, err)
	}
}

func Run(fn RunFunc, ctxs ...context.Context) Task {
	ctx := firstContext(ctxs, nil)
	if ctx != nil && isCanceledContext(ctx) {
		return Cancel(ctx.Err())
	}
	return newStarter(runCaller(fn), ctx)
}

/* Delay */
func Delay(d time.Duration, ctxs ...context.Context) Task {
	ctx := firstContext(ctxs, nil)
	if ctx == nil {
		ctx = context.TODO()
	} else if isCanceledContext(ctx) {
		return Cancel(ctx.Err())
	}
	var fn taskCaller = func(task *TaskImpl) {
		timer := time.NewTimer(d)
		defer timer.Stop()
		select {
		case <-timer.C:
			resolve(task, nil)
		case <-ctx.Done():
			cancel(task, ctx.Err())
		}
	}
	return newStarter(fn, ctx)
}

func waitAllTask(tasks []Task, ctxs ...context.Context) {
	c := 0
	for _, task := range tasks {
		if task != nil && !task.IsDone() {
			c++
		}
	}
	if c == 0 {
		return
	}
	ctx := firstContext(ctxs, context.TODO())
	wc := make(chan struct{}, c)
	defer close(wc)

	watiFn := func(Task) (interface{}, error) {
		select {
		case wc <- struct{}{}:
		default:
		}
		return nil, nil
	}

	for _, task := range tasks {
		if task != nil && !task.IsDone() {
			task.Continue(watiFn, ctx)
		}
	}
	for i := 0; i < c; i++ {
		select {
		case <-ctx.Done():
		case _, ok := <-wc:
			if ok {
				continue
			}
		}
		return
	}

}

func WaitAll(tasks ...Task) {
	waitAllTask(tasks, nil)
}
func WaitAllWithContext(tasks ...Task) func(ctxs ...context.Context) {
	return func(ctxs ...context.Context) {
		waitAllTask(tasks, ctxs...)
	}
}

func waitAnyTask(tasks []Task, ctxs ...context.Context) {
	for _, task := range tasks {
		if task != nil && task.IsDone() {
			return
		}
	}

	ctx := firstContext(ctxs, context.TODO())
	wc := make(chan struct{})
	defer close(wc)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	watiFn := func(Task) (interface{}, error) {
		select {
		case wc <- struct{}{}:
		default:
		}
		return nil, nil
	}

	for _, task := range tasks {
		if task != nil {
			task.Continue(watiFn, ctx)
		}
	}
	select {
	case <-ctx.Done():
	case <-wc:
	}
}

func WaitAny(tasks ...Task) {
	waitAnyTask(tasks, nil)
}
func WaitAnyWithContext(tasks ...Task) func(ctxs ...context.Context) {
	return func(ctxs ...context.Context) {
		waitAnyTask(tasks, ctxs...)
	}
}
