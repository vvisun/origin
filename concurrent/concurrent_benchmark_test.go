package concurrent

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkAsyncDo 基准测试：异步执行任务
func BenchmarkAsyncDo(b *testing.B) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 10, 100000)
	defer c.Close()

	var completed int32
	done := make(chan bool)

	// 启动回调处理goroutine
	go func() {
		for {
			select {
			case cb := <-c.GetCallBackChannel():
				c.DoCallback(cb)
				if atomic.AddInt32(&completed, 1) >= int32(b.N) {
					close(done)
					return
				}
			case <-done:
				return
			}
		}
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.AsyncDo(func() bool {
				return true
			}, func(err error) {
				// 回调处理在单独的goroutine中
			})
		}
	})

	// 等待所有任务完成
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		b.Fatal("timeout waiting for tasks")
	}
}

// BenchmarkAsyncDoByQueue 基准测试：队列任务
func BenchmarkAsyncDoByQueue(b *testing.B) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 10, 100000)
	defer c.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		queueId := int64(1)
		for pb.Next() {
			c.AsyncDoByQueue(queueId, func() bool {
				return true
			}, nil)
		}
	})
}

// BenchmarkConcurrentHighLoad 基准测试：高负载场景
func BenchmarkConcurrentHighLoad(b *testing.B) {
	c := &Concurrent{}
	c.OpenConcurrent(10, 50, 1000000)
	defer c.Close()

	var completed int32
	done := make(chan bool)

	// 启动回调处理goroutine
	go func() {
		for {
			select {
			case cb := <-c.GetCallBackChannel():
				c.DoCallback(cb)
				if atomic.AddInt32(&completed, 1) >= int32(b.N) {
					close(done)
					return
				}
			case <-done:
				return
			}
		}
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.AsyncDo(func() bool {
				// 模拟一些工作
				time.Sleep(1 * time.Microsecond)
				return true
			}, func(err error) {
				// 回调处理
			})
		}
	})

	select {
	case <-done:
	case <-time.After(60 * time.Second):
		b.Fatal("timeout waiting for tasks")
	}
}

// BenchmarkQueueOperations 基准测试：队列操作
func BenchmarkQueueOperations(b *testing.B) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 10, 100000)
	defer c.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queueId := int64(i % 100) // 使用100个不同的队列
		c.AsyncDoByQueue(queueId, func() bool {
			return true
		}, nil)
	}
}

// BenchmarkWorkerCreation 基准测试：worker创建
func BenchmarkWorkerCreation(b *testing.B) {
	c := &Concurrent{}
	c.OpenConcurrent(1, 100, 100000)
	defer c.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.AsyncDo(func() bool {
				time.Sleep(10 * time.Microsecond)
				return true
			}, nil)
		}
	})
}

// BenchmarkConcurrentMixed 基准测试：混合任务类型
func BenchmarkConcurrentMixed(b *testing.B) {
	c := &Concurrent{}
	c.OpenConcurrent(5, 20, 100000)
	defer c.Close()

	var completed int32
	done := make(chan bool)

	go func() {
		for {
			select {
			case cb := <-c.GetCallBackChannel():
				c.DoCallback(cb)
				if atomic.AddInt32(&completed, 1) >= int32(b.N) {
					close(done)
					return
				}
			case <-done:
				return
			}
		}
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := int64(0)
		for pb.Next() {
			if i%2 == 0 {
				// 普通任务
				c.AsyncDo(func() bool {
					return true
				}, func(err error) {})
			} else {
				// 队列任务
				c.AsyncDoByQueue(i%10, func() bool {
					return true
				}, func(err error) {})
			}
			i++
		}
	})

	select {
	case <-done:
	case <-time.After(60 * time.Second):
		b.Fatal("timeout")
	}
}

// BenchmarkCallbackProcessing 基准测试：回调处理
func BenchmarkCallbackProcessing(b *testing.B) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 10, 100000)
	defer c.Close()

	var wg sync.WaitGroup
	var completed int32
	wg.Add(b.N)

	// 处理回调 - 确保处理所有回调
	done := make(chan bool)
	go func() {
		defer close(done)
		timeout := time.After(30 * time.Second)
		for atomic.LoadInt32(&completed) < int32(b.N) {
			select {
			case cb := <-c.GetCallBackChannel():
				c.DoCallback(cb)
			case <-timeout:
				return
			}
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.AsyncDo(func() bool {
			return true
		}, func(err error) {
			atomic.AddInt32(&completed, 1)
			wg.Done()
		})
	}

	// 等待所有回调完成
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		b.Fatal("timeout waiting for callbacks")
	}

	wg.Wait()
}

// BenchmarkConcurrentStress 压力测试：持续高负载
func BenchmarkConcurrentStress(b *testing.B) {
	c := &Concurrent{}
	c.OpenConcurrent(10, 100, 1000000)
	defer c.Close()

	const numGoroutines = 100
	var wg sync.WaitGroup
	stop := make(chan bool)

	// 启动多个goroutine持续提交任务
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					c.AsyncDo(func() bool {
						return true
					}, nil)
					time.Sleep(1 * time.Microsecond)
				}
			}
		}()
	}

	// 运行指定时间
	time.Sleep(time.Duration(b.N) * time.Millisecond)

	// 停止所有goroutine
	close(stop)
	wg.Wait()
}
