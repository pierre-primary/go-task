package task_test

import (
	"sync"
	"testing"

	"github.com/pierre-primary/go-task/test/lock"
)

func Benchmark_Mutex(b *testing.B) {
	var lock = sync.Mutex{}
	cpu := 100
	v := int64(0)
	var wg sync.WaitGroup
	wg.Add(cpu)
	b.ResetTimer()
	for i := 0; i < cpu; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < b.N; j++ {
				lock.Lock()
				v++
				lock.Unlock()
			}
		}()
	}
	wg.Wait()
	b.StopTimer()
	if v != int64(b.N)*int64(cpu) {
		b.Log("xxx")
	}
}

func Benchmark_ChanLock(b *testing.B) {
	var semaphore = make(chan bool, 1)
	cpu := 100
	v := int64(0)
	var wg sync.WaitGroup
	wg.Add(cpu)
	b.ResetTimer()
	for i := 0; i < cpu; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < b.N; j++ {
				semaphore <- true
				v++
				<-semaphore
			}
		}()
	}
	wg.Wait()
	b.StopTimer()
	if v != int64(b.N)*int64(cpu) {
		b.Log("xxx")
	}
}

func Benchmark_SpinLock(b *testing.B) {
	var lock lock.SpinLock
	cpu := 100
	v := int64(0)
	var wg sync.WaitGroup
	wg.Add(cpu)
	b.ResetTimer()
	for i := 0; i < cpu; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < b.N; j++ {
				lock.Lock()
				v++
				lock.Unlock()
			}
		}()
	}
	wg.Wait()
	b.StopTimer()
	if v != int64(b.N)*int64(cpu) {
		b.Log("xxx")
	}
}

func Benchmark_RWMutex(b *testing.B) {
	var lock sync.RWMutex
	cpu := 100
	v := int64(0)
	for i := 0; i < cpu; i++ {
		go func() {
			for j := 0; j < b.N; j++ {
				lock.RLock()
				_ = v
				lock.RUnlock()
			}
		}()
	}
	var wg sync.WaitGroup
	wg.Add(cpu)
	b.ResetTimer()
	for i := 0; i < cpu; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < b.N; j++ {
				lock.Lock()
				v++
				lock.Unlock()
			}
		}()
	}
	wg.Wait()
	b.StopTimer()
	if v != int64(b.N)*int64(cpu) {
		b.Log("xxx")
	}
}

func Benchmark_SpinRWLock(b *testing.B) {
	var lock lock.SpinRWLock
	cpu := 100
	v := int64(0)
	for i := 0; i < cpu; i++ {
		go func() {
			for j := 0; j < b.N; j++ {
				lock.RLock()
				_ = v
				lock.RUnlock()
			}
		}()
	}
	var wg sync.WaitGroup
	wg.Add(cpu)
	b.ResetTimer()
	for i := 0; i < cpu; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < b.N; j++ {
				lock.Lock()
				v++
				lock.Unlock()
			}
		}()
	}
	wg.Wait()
	b.StopTimer()
	if v != int64(b.N)*int64(cpu) {
		b.Log("xxx")
	}
}
