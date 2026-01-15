package concurrent

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestOpenConcurrent 测试打开并发池
func TestOpenConcurrent(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 5, 100)
	defer c.Close()

	if cap(c.tasks) != 100 {
		t.Errorf("expected tasks channel capacity 100, got %d", cap(c.tasks))
	}
}

// TestOpenConcurrentByNumCPU 测试基于CPU数量打开并发池
func TestOpenConcurrentByNumCPU(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrentByNumCPU(1.0)
	defer c.Close()

	if cap(c.tasks) == 0 {
		t.Error("expected tasks channel to be initialized")
	}
}

// TestOpenConcurrentPanic 测试重复打开应该panic
func TestOpenConcurrentPanic(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 5, 100)
	defer c.Close()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on repeated OpenConcurrent call")
		}
	}()

	c.OpenConcurrent(2, 5, 100)
}

// TestAsyncDo 测试异步执行任务
func TestAsyncDo(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 5, 100)
	defer c.Close()

	var executed int32
	var callbackExecuted int32

	// 执行任务
	c.AsyncDo(func() bool {
		atomic.AddInt32(&executed, 1)
		return true
	}, func(err error) {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		atomic.AddInt32(&callbackExecuted, 1)
	})

	// 等待回调执行
	timeout := time.After(2 * time.Second)
	for atomic.LoadInt32(&callbackExecuted) == 0 {
		select {
		case cb := <-c.GetCallBackChannel():
			c.DoCallback(cb)
		case <-timeout:
			t.Fatal("timeout waiting for callback")
		}
	}

	if atomic.LoadInt32(&executed) != 1 {
		t.Errorf("expected task to execute once, got %d", atomic.LoadInt32(&executed))
	}

	if atomic.LoadInt32(&callbackExecuted) != 1 {
		t.Errorf("expected callback to execute once, got %d", atomic.LoadInt32(&callbackExecuted))
	}
}

// TestAsyncDoNoCallback 测试没有回调的任务
func TestAsyncDoNoCallback(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 5, 100)
	defer c.Close()

	var executed int32

	c.AsyncDo(func() bool {
		atomic.AddInt32(&executed, 1)
		return true
	}, nil)

	// 等待任务执行
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&executed) != 1 {
		t.Errorf("expected task to execute once, got %d", atomic.LoadInt32(&executed))
	}
}

// TestAsyncDoReturnFalse 测试返回false时不执行回调
func TestAsyncDoReturnFalse(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 5, 100)
	defer c.Close()

	var executed int32
	var callbackExecuted int32

	c.AsyncDo(func() bool {
		atomic.AddInt32(&executed, 1)
		return false // 返回false，不应该执行回调
	}, func(err error) {
		atomic.AddInt32(&callbackExecuted, 1)
	})

	// 等待任务执行
	time.Sleep(100 * time.Millisecond)

	// 处理可能的回调
	select {
	case cb := <-c.GetCallBackChannel():
		c.DoCallback(cb)
	case <-time.After(100 * time.Millisecond):
		// 没有回调是正常的
	}

	if atomic.LoadInt32(&executed) != 1 {
		t.Errorf("expected task to execute once, got %d", atomic.LoadInt32(&executed))
	}

	if atomic.LoadInt32(&callbackExecuted) != 0 {
		t.Errorf("expected callback not to execute, got %d", atomic.LoadInt32(&callbackExecuted))
	}
}

// TestAsyncDoByQueue 测试队列任务
func TestAsyncDoByQueue(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 5, 100)
	defer c.Close()

	var executionOrder []int32
	var mu sync.Mutex

	queueId := int64(1)

	// 提交多个任务到同一个队列
	for i := int32(0); i < 5; i++ {
		taskNum := i
		c.AsyncDoByQueue(queueId, func() bool {
			mu.Lock()
			executionOrder = append(executionOrder, taskNum)
			mu.Unlock()
			time.Sleep(10 * time.Millisecond) // 模拟工作
			return true
		}, nil)
	}

	// 等待所有任务完成
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if len(executionOrder) != 5 {
		t.Errorf("expected 5 tasks to execute, got %d", len(executionOrder))
	}

	// 验证任务按顺序执行（队列任务应该串行执行）
	for i := 0; i < len(executionOrder)-1; i++ {
		if executionOrder[i] >= executionOrder[i+1] {
			t.Errorf("tasks should execute in order, got %v", executionOrder)
			break
		}
	}
	mu.Unlock()
}

// TestAsyncDoCallbackOnly 测试只有回调的情况
func TestAsyncDoCallbackOnly(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 5, 100)
	defer c.Close()

	var callbackExecuted int32

	c.AsyncDoByQueue(0, nil, func(err error) {
		atomic.AddInt32(&callbackExecuted, 1)
	})

	// 等待回调执行
	timeout := time.After(2 * time.Second)
	for atomic.LoadInt32(&callbackExecuted) == 0 {
		select {
		case cb := <-c.GetCallBackChannel():
			c.DoCallback(cb)
		case <-timeout:
			t.Fatal("timeout waiting for callback")
		}
	}

	if atomic.LoadInt32(&callbackExecuted) != 1 {
		t.Errorf("expected callback to execute once, got %d", atomic.LoadInt32(&callbackExecuted))
	}
}

// TestAsyncDoError 测试任务执行错误
func TestAsyncDoError(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 5, 100)
	defer c.Close()

	var callbackError error

	c.AsyncDo(func() bool {
		panic("test error")
	}, func(err error) {
		callbackError = err
	})

	// 等待回调执行
	timeout := time.After(2 * time.Second)
	for callbackError == nil {
		select {
		case cb := <-c.GetCallBackChannel():
			c.DoCallback(cb)
		case <-timeout:
			t.Fatal("timeout waiting for callback")
		}
	}

	if callbackError == nil {
		t.Error("expected error in callback")
	}
}

// TestClose 测试关闭并发池
func TestClose(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 5, 100)

	// 提交一些任务
	for i := 0; i < 10; i++ {
		c.AsyncDo(func() bool {
			time.Sleep(10 * time.Millisecond)
			return true
		}, nil)
	}

	// 关闭
	done := make(chan bool)
	go func() {
		c.Close()
		close(done)
	}()

	select {
	case <-done:
		// 关闭成功
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for close")
	}
}

// TestCloseMultipleTimes 测试多次关闭
func TestCloseMultipleTimes(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 5, 100)

	c.Close()
	c.Close() // 应该可以安全地多次调用
}

// TestChannelFull 测试任务通道满的情况
func TestChannelFull(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 5, 2) // 小容量通道
	defer c.Close()

	var errorReceived int32
	var wg sync.WaitGroup

	// 启动回调处理goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		timeout := time.After(3 * time.Second)
		for {
			select {
			case cb := <-c.GetCallBackChannel():
				c.DoCallback(cb)
			case <-timeout:
				return
			}
		}
	}()

	// 填满通道
	for i := 0; i < 10; i++ {
		c.AsyncDoByQueue(0, func() bool {
			time.Sleep(10 * time.Millisecond) // 减少睡眠时间
			return true
		}, func(err error) {
			if err != nil {
				atomic.AddInt32(&errorReceived, 1)
			}
		})
	}

	// 等待一些任务完成
	time.Sleep(500 * time.Millisecond)

	// 等待回调处理完成
	wg.Wait()
}

// TestWorkerNumLimit 测试worker数量限制
func TestWorkerNumLimit(t *testing.T) {
	c := &Concurrent{}
	maxWorkers := int32(3)
	c.OpenConcurrent(1, maxWorkers, 100)
	defer c.Close()

	var activeWorkers int32
	var mu sync.Mutex

	// 提交大量任务，每个任务都会阻塞一段时间
	for i := 0; i < 20; i++ {
		c.AsyncDo(func() bool {
			mu.Lock()
			current := atomic.AddInt32(&activeWorkers, 1)
			mu.Unlock()

			if current > maxWorkers {
				t.Errorf("worker count exceeded max: %d > %d", current, maxWorkers)
			}

			time.Sleep(50 * time.Millisecond)

			mu.Lock()
			atomic.AddInt32(&activeWorkers, -1)
			mu.Unlock()

			return true
		}, nil)
	}

	// 等待所有任务完成
	time.Sleep(2 * time.Second)
}

// TestQueueIdMapping 测试queueId映射
func TestQueueIdMapping(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 5, 100)
	defer c.Close()

	var executionOrder []int64
	var mu sync.Mutex

	// 使用不同的queueId，但会被映射到同一个队列
	queueId1 := int64(1)
	queueId2 := int64(maxTaskQueueSessionId + 1) // 会被映射到1

	c.AsyncDoByQueue(queueId1, func() bool {
		mu.Lock()
		executionOrder = append(executionOrder, queueId1)
		mu.Unlock()
		time.Sleep(20 * time.Millisecond)
		return true
	}, nil)

	c.AsyncDoByQueue(queueId2, func() bool {
		mu.Lock()
		executionOrder = append(executionOrder, queueId2)
		mu.Unlock()
		time.Sleep(20 * time.Millisecond)
		return true
	}, nil)

	// 等待任务完成
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if len(executionOrder) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(executionOrder))
	}
	mu.Unlock()
}

// TestNilFnAndCb 测试fn和cb都为nil的情况
func TestNilFnAndCb(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 5, 100)
	defer c.Close()

	// 应该不会panic，只是记录错误
	c.AsyncDoByQueue(0, nil, nil)
	time.Sleep(10 * time.Millisecond)
}

// TestDoCallbackPanic 测试回调panic恢复
func TestDoCallbackPanic(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 5, 100)
	defer c.Close()

	// 创建一个会panic的回调
	cb := func(err error) {
		panic("callback panic")
	}

	// 应该不会导致程序崩溃
	c.DoCallback(cb)
}

// TestConcurrentBasic 测试基本并发场景
func TestConcurrentBasic(t *testing.T) {
	c := &Concurrent{}
	c.OpenConcurrent(2, 10, 1000)
	defer c.Close()

	const numTasks = 100
	var completed int32
	var wg sync.WaitGroup

	wg.Add(numTasks)

	for i := 0; i < numTasks; i++ {
		c.AsyncDo(func() bool {
			return true
		}, func(err error) {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			atomic.AddInt32(&completed, 1)
			wg.Done()
		})
	}

	// 处理回调
	done := make(chan bool)
	go func() {
		for atomic.LoadInt32(&completed) < numTasks {
			select {
			case cb := <-c.GetCallBackChannel():
				c.DoCallback(cb)
			case <-time.After(5 * time.Second):
				t.Error("timeout processing callbacks")
				return
			}
		}
		close(done)
	}()

	select {
	case <-done:
		if atomic.LoadInt32(&completed) != numTasks {
			t.Errorf("expected %d tasks completed, got %d", numTasks, atomic.LoadInt32(&completed))
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for all tasks")
	}

	wg.Wait()
}
