
package task_test

import (
	"sync"
	"testing"

	"github.com/pierre-primary/go-task"
)

type any = interface{}

func Benchmark_Task(b *testing.B) {
	var t task.Task
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t = task.Run(func() (any, error) {
			return 1, nil
		}).Then(func(i any) (any, error) {
			return 2, nil
		}).Then(func(i any) (interface{}, error) {
			return 3, nil
		}).Wait()
	}
	b.StopTimer()
	_ = t
}

func Benchmark_Normal(b *testing.B) {
	var v any
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := make(chan any)
		go func() {
			_ = v
			ch <- 1
		}()
		v = <-ch
		go func() {
			_ = v
			ch <- 2
		}()
		v = <-ch
		go func() {
			_ = v
			ch <- 3
		}()
		v = <-ch
	}
	b.StopTimer()
	_ = v
}

func Benchmark_WaitGourp(b *testing.B) {
	var v any
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = v
			v = 1
		}()
		wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = v
			v = 2
		}()
		wg.Wait()
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = v
			v = 3
		}()
		wg.Wait()
	}
	b.StopTimer()
	_ = v
}
