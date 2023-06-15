package task_test

import (
	"testing"
)

func fn1(a int, b int) func() {
	return func() {

	}
}

func Benchmark_fn1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fn1(i, i)()
	}
}

func fn2(a int, b int) func() int {
	return func() int {
		return a + b
	}
}

func Benchmark_fn2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fn2(i, i)()
	}
}

func fn3(a int, b int) func() int {
	return func() int {
		a += b
		return a
	}
}

func Benchmark_fn3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fn3(i, i)()
	}
}

func fn4(a int, b int) func() int {
	return func() int {
		a := a
		b := b
		a += b
		return a
	}
}

func Benchmark_fn4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fn4(i, i)()
	}
}

func fn5(a int, b int) func() int {
	st := &struct{ a, b int }{a: a, b: b}
	return func() int {
		st.a += st.b
		return st.a
	}
}

func Benchmark_fn5(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fn5(i, i)()
	}
}
