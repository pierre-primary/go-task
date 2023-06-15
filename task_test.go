package task_test

import (
	"errors"
	"testing"

	"github.com/pierre-primary/go-task"
)

func Test_New(t *testing.T) {
	t.Run("Resolve", func(t *testing.T) {
		task, resolve, _ := task.New()
		task = task.Then(func(rs any) (any, error) {
			t.Log(rs)
			return nil, nil
		}).Catch(func(err error) (any, error) {
			t.Error("错误的路径")
			return nil, nil
		})
		resolve("Resolve")
		task.Wait()
	})
	t.Run("Reject", func(t *testing.T) {
		task, _, reject := task.New()
		task = task.Then(func(rs any) (any, error) {
			t.Error("错误的路径")
			return nil, nil
		}).Catch(func(err error) (any, error) {
			t.Log(err)
			return nil, nil
		})
		reject(errors.New("Reject"))
		task.Wait()
	})
}

func Test_Start(t *testing.T) {
	t.Run("Resolve", func(t *testing.T) {
		task.Start(func(resolve task.ResolveFunc, _ task.RejectFunc) {
			resolve("Resolve")
		}).Then(func(rs any) (any, error) {
			t.Log(rs)
			return nil, nil
		}).Catch(func(err error) (any, error) {
			t.Error("错误的路径")
			return nil, nil
		}).Wait()
	})
	t.Run("Reject", func(t *testing.T) {
		task.Start(func(_ task.ResolveFunc, reject task.RejectFunc) {
			reject(errors.New("Reject"))
		}).Then(func(rs any) (any, error) {
			t.Error("错误的路径")
			return nil, nil
		}).Catch(func(err error) (any, error) {
			t.Log(err)
			return nil, nil
		}).Wait()
	})
}
func Test_Run(t *testing.T) {
	t.Run("Resolve", func(t *testing.T) {
		task.Run(func() (any, error) {
			return "Resolve", nil
		}).Wait()
	})
}

func Test_Done(t *testing.T) {
	t.Run("Resolve", func(t *testing.T) {
		task.Resolve()
		task.Resolve(nil)
		task.Resolve(1)
	})
	t.Run("Reject", func(t *testing.T) {
		task.Reject()
		task.Reject(nil)
		task.Reject(errors.New("Reject"))
	})
	t.Run("Cancel", func(t *testing.T) {
		task.Cancel()
		task.Cancel(nil)
		task.Cancel(errors.New("Cancel"))
	})
}

func Test_Error(t *testing.T) {
	task.Run(func() (interface{}, error) {
		panic("xxx")
	}).Wait()
}
