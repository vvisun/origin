package concurrent

import (
	"errors"
	"runtime"

	"sync/atomic"

	"github.com/duanhf2012/origin/v2/log"
)

const defaultMaxTaskChannelNum = 1000000

type IConcurrent interface {
	//cpuMul 表示cpu的倍数。建议:(1)cpu密集型 使用1  (2)i/o密集型使用2或者更高
	OpenConcurrentByNumCPU(cpuMul float32)
	// minGoroutineNum 表示最小协程数
	// maxGoroutineNum 表示最大协程数
	// maxTaskChannelNum 表示任务管道的大小
	OpenConcurrent(minGoroutineNum int32, maxGoroutineNum int32, maxTaskChannelNum int)
	// 参数一传入队列Id, 同一个队列Id将在协程池中被排队执行
	AsyncDoByQueue(queueId int64, fn func() bool, cb func(err error))
	// 参数一的函数在其他协程池中执行完成，将执行完成事件放入服务工作协程，
	// 参数二的函数在服务协程中执行，是协程安全的。
	// 函数参数可以某中一个为空。 f为空时cb函数将被延迟执行; cb为空时，f在协程池中执行，但没有在服务协程中回调
	// 参数一返回false时，cb函数将不会被执行, 为true时，则会被执行
	AsyncDo(f func() bool, cb func(err error))
}

type Concurrent struct {
	dispatch

	tasks     chan task
	cbChannel chan func(error)
	open      int32
}

/*
	OpenConcurrentByNumCPU 函数使用说明

cpuMul 表示cpu的倍数
建议:(1)cpu密集型 使用1  (2)i/o密集型使用2或者更高
*/
func (c *Concurrent) OpenConcurrentByNumCPU(cpuNumMul float32) {
	goroutineNum := int32(float32(runtime.NumCPU())*cpuNumMul + 1)
	c.OpenConcurrent(goroutineNum, goroutineNum, defaultMaxTaskChannelNum)
}

func (c *Concurrent) OpenConcurrent(minGoroutineNum int32, maxGoroutineNum int32, maxTaskChannelNum int) {
	if atomic.AddInt32(&c.open, 1) > 1 {
		panic("repeated calls to OpenConcurrent are not allowed!")
	}

	c.tasks = make(chan task, maxTaskChannelNum)
	c.cbChannel = make(chan func(error), maxTaskChannelNum)

	//打开dispach
	c.dispatch.open(minGoroutineNum, maxGoroutineNum, c.tasks, c.cbChannel)
}

func (c *Concurrent) AsyncDo(f func() bool, cb func(err error)) {
	c.AsyncDoByQueue(0, f, cb)
}

func (c *Concurrent) AsyncDoByQueue(queueId int64, fn func() bool, cb func(err error)) {
	if cap(c.tasks) == 0 {
		panic("not open concurrent")
	}

	if fn == nil && cb == nil {
		log.StackError("fn and cb is nil")
		return
	}

	if fn == nil {
		c.pushAsyncDoCallbackEvent(cb)
		return
	}

	if queueId != 0 {
		queueId = queueId%maxTaskQueueSessionId + 1
	}

	select {
	case c.tasks <- task{queueId, fn, cb}:
	default:
		log.Error("tasks channel is full")
		if cb != nil {
			c.pushAsyncDoCallbackEvent(func(err error) {
				cb(errors.New("tasks channel is full"))
			})
		}
		return
	}
}

func (c *Concurrent) Close() {
	if cap(c.tasks) == 0 {
		return
	}

	log.Info("wait close concurrent")

	c.dispatch.close()

	log.Info("concurrent has successfully exited")
}

func (c *Concurrent) GetCallBackChannel() chan func(error) {
	return c.cbChannel
}
