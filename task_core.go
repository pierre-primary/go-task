package task

import (
	"runtime"
	"sync/atomic"
)

const (
	/* state bit */
	flagCompleted  uint32 = 0b0001                                    // 任务完成
	flagFailed     uint32 = 0b0010                                    // 任务失败
	flagCanceled   uint32 = 0b0100                                    // 任务取消
	flagWaitSub    uint32 = 0b1000                                    // 等待子任务
	maskState      uint32 = (1 << 12) - 1                             // 状态标记位掩码
	checkCompleted uint32 = flagCompleted                             // 完成状态掩码
	checkDone      uint32 = flagCompleted | flagFailed | flagCanceled // 结束状态掩码（完成，失败，取消）

	/* lock bit */
	lockFlws  uint32 = 0b0001 << 12         // Follower 链表相关操作的锁标记位
	lockChan  uint32 = 0b0010 << 12         // 通道相关操作的锁标记位
	lockState uint32 = lockFlws | lockChan  // 全锁
	maskLock  uint32 = ((1 << 4) - 1) << 12 // 锁标记位掩码

	/* option bit */
	// nolint:unused
	maskOptions uint32 = ((1 << 16) - 1) << 16 // 选项标记位掩码
)

type (
	TaskImpl struct {
		// |———————— option（16 bit）————————|-- lock（4 bit）--|—————— state（12 bit）——————|
		state uint32
		data  interface{}
		flws  *Follower
		ch    chan struct{}
	}
)

var closedChan = make(chan struct{})

func init() {
	close(closedChan)
}

// ------------------------------------------------------------------------------------------------------
/* Task State Manage */

// Load the state
func state(task *TaskImpl) uint32 {
	return atomic.LoadUint32(&task.state)
}

// Check the state
func stateIs(task *TaskImpl, state uint32) bool {
	return atomic.LoadUint32(&task.state)&state != 0
}

// Set the state with condition
func setStateIfNot(task *TaskImpl, stateFlags, checkFlags uint32) bool {
	for {
		state := atomic.LoadUint32(&task.state)
		if state&checkFlags != 0 {
			// does not meet set state conditions.
			return false
		} else if atomic.CompareAndSwapUint32(&task.state, state, (state&^maskState)|stateFlags) {
			// successfully set the state.
			return true
		}
	}
}

// Lock with condition
func lockStateIfNot(task *TaskImpl, lockFlags uint32, checkFlags uint32) bool {
	for {
		state := atomic.LoadUint32(&task.state)
		if state&checkFlags != 0 {
			// does not meet locking conditions.
			return false
		} else if state&lockFlags != 0 {
			// lock held by another goroutine.
			runtime.Gosched()
		} else if atomic.CompareAndSwapUint32(&task.state, state, state|lockFlags) {
			// successfully locked.
			return true
		}
	}
}

// Unlock and set state
func unlockStateAndSet(task *TaskImpl, clearFlags uint32, stateFlags uint32) {
	if stateFlags != 0 {
		// if you need to set the new state,
		// you need to clear the old state flags first.
		clearFlags |= maskState
	}
	for {
		state := atomic.LoadUint32(&task.state)
		if atomic.CompareAndSwapUint32(&task.state, state, (state&^clearFlags)|stateFlags) {
			break
		}
	}
}

// Get a channel to wait for task done
func done(task *TaskImpl) chan struct{} {
	// check and lock
	if lockStateIfNot(task, lockChan, checkDone) {
		ch := make(chan struct{})
		task.ch = ch
		unlockStateAndSet(task, lockChan, 0)
		return ch
	} else {
		return task.ch
	}
}

// Synchronously wait for task done
func wait(task *TaskImpl) {
	if stateIs(task, checkDone) {
		return
	}
	<-done(task)
}

// ------------------------------------------------------------------------------------------------------
/* Task Core */

// Create an initial task
func newTask() *TaskImpl {
	return &TaskImpl{state: 0}
}

// Create a done task
func newDoneTask(flags uint32, data interface{}) *TaskImpl {
	return &TaskImpl{state: flags, data: data, ch: closedChan}
}

// Terminate task
func terminate(task *TaskImpl, state uint32, data interface{}) {
	// check and lock
	if lockStateIfNot(task, lockState, checkDone) {
		// temporarily pending time-consuming operations
		tempCh := task.ch

		task.data = data
		task.ch = closedChan

		// unlock
		unlockStateAndSet(task, lockState, state)

		// handle time-consuming operations
		if tempCh != nil {
			close(tempCh)
		}
		wakeAllFollower(task)
	}
}

// Resolve task
func resolve(task *TaskImpl, result interface{}) {
	// check sub task
	if sub, ok := result.(Task); ok {
		// wait task
		if setStateIfNot(task, flagWaitSub, checkDone|flagWaitSub) {
			if task == sub {
				internalPanicForce("A task cannot be resolved with itself.")
				return
			}
			sub.Wait()
			// copy the result
			switch sub.State() {
			case STATE_COMPLETED:
				resolve(task, sub.Result())
			case STATE_CANCELED:
				cancel(task, sub.Error())
			default:
				reject(task, sub.Error())
			}
		}
	} else {
		terminate(task, flagCompleted, result)
	}
}

// Reject task
func reject(task *TaskImpl, msg interface{}) {
	switch v := msg.(type) {
	case *ForcePanic:
		panic(v.msg)
	case error:
		terminate(task, flagFailed, v)
	default:
		terminate(task, flagFailed, toError(msg))
	}
}

// Cancel task
func cancel(task *TaskImpl, err error) {
	terminate(task, flagCanceled, err)
}
