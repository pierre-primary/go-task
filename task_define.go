package task

import (
	"context"
	"time"
)

type TaskState = uint32

// 状态定义
const (
	STATE_PENDING   TaskState = 0 // 任务等待
	STATE_COMPLETED TaskState = 1 // 任务完成
	STATE_FAULTED   TaskState = 2 // 任务故障
	STATE_CANCELED  TaskState = 3 // 任务取消
)

// 回调函数定义
type (
	ContinueFunc = func(Task) (interface{}, error)
	ThenFunc     = func(interface{}) (interface{}, error)
	CatchFunc    = func(error) (interface{}, error)
)

// 任务接口定义
type Task interface {
	Continue(ContinueFunc, ...context.Context) Task
	ContinueAwait(ContinueFunc, ...context.Context) Task
	Then(ThenFunc, ...context.Context) Task
	ThenAwait(ThenFunc, ...context.Context) Task
	Catch(CatchFunc, ...context.Context) Task
	CatchAwait(CatchFunc, ...context.Context) Task
	Done() chan struct{}
	Wait(ctxs ...context.Context) Task
	WaitTimeout(time.Duration, ...context.Context) Task
	State() TaskState
	IsDone() bool
	IsCompleted() bool
	IsFaulted() bool
	IsCanceled() bool
	Return() (interface{}, error)
	Result() interface{}
	Error() error
}
