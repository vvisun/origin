package concurrent

import (
	"sync"
	"sync/atomic"
	"time"

	"fmt"

	"context"

	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/util/queue"
)

var idleTimeout = int64(2 * time.Second)

const maxTaskQueueSessionId = 10000

type dispatch struct {
	minConcurrentNum int32
	maxConcurrentNum int32

	queueIdChannel chan int64
	workerQueue    chan task
	tasks          chan task
	idle           int32 // 使用原子操作保护
	workerNum      int32 // 使用原子操作保护
	cbChannel      chan func(error)

	mapTaskQueueSession map[int64]*queue.Deque[task]
	queueSessionMu      sync.RWMutex // 保护 mapTaskQueueSession 的并发访问

	waitWorker   sync.WaitGroup
	waitDispatch sync.WaitGroup

	cancelContext context.Context
	cancel        context.CancelFunc
}

func (d *dispatch) open(minGoroutineNum int32, maxGoroutineNum int32, tasks chan task, cbChannel chan func(error)) {
	d.minConcurrentNum = minGoroutineNum
	d.maxConcurrentNum = maxGoroutineNum
	d.tasks = tasks
	d.mapTaskQueueSession = make(map[int64]*queue.Deque[task], maxTaskQueueSessionId)
	d.workerQueue = make(chan task)
	d.cbChannel = cbChannel
	d.queueIdChannel = make(chan int64, cap(tasks))
	d.cancelContext, d.cancel = context.WithCancel(context.Background())
	// 初始化 idle 为 true（使用原子操作）
	atomic.StoreInt32(&d.idle, 1)
	d.waitDispatch.Add(1)
	go d.run()
}

func (d *dispatch) run() {
	defer d.waitDispatch.Done()
	timeout := time.NewTimer(time.Duration(atomic.LoadInt64(&idleTimeout)))

	for {
		select {
		case queueId := <-d.queueIdChannel:
			d.processQueueEvent(queueId)
		default:
			select {
			case t, ok := <-d.tasks:
				if ok == false {
					return
				}
				d.processTask(&t)
			case queueId := <-d.queueIdChannel:
				d.processQueueEvent(queueId)
			case <-timeout.C:
				d.processTimer()
				// 修复：重置定时器
				timeout.Reset(time.Duration(atomic.LoadInt64(&idleTimeout)))
			case <-d.cancelContext.Done():
				atomic.StoreInt64(&idleTimeout, int64(time.Millisecond*5))
				timeout.Reset(time.Duration(atomic.LoadInt64(&idleTimeout)))
				workerNum := atomic.LoadInt32(&d.workerNum)
				for i := int32(0); i < workerNum; i++ {
					d.processIdle()
				}
			}
		}

		if atomic.LoadInt32(&d.minConcurrentNum) == -1 && atomic.LoadInt32(&d.workerNum) == 0 {
			d.waitWorker.Wait()
			d.cbChannel <- nil
			return
		}
	}
}

func (d *dispatch) processTimer() {
	// 使用原子操作读取 idle 和 workerNum
	if atomic.LoadInt32(&d.idle) == 1 && atomic.LoadInt32(&d.workerNum) > atomic.LoadInt32(&d.minConcurrentNum) {
		d.processIdle()
	}

	atomic.StoreInt32(&d.idle, 1)
}

func (d *dispatch) processQueueEvent(queueId int64) {
	atomic.StoreInt32(&d.idle, 0)

	d.queueSessionMu.RLock()
	queueSession := d.mapTaskQueueSession[queueId]
	d.queueSessionMu.RUnlock()

	if queueSession == nil {
		return
	}

	queueSession.PopFront()
	if queueSession.Len() == 0 {
		// 修复：清理空队列，防止内存泄漏
		d.queueSessionMu.Lock()
		delete(d.mapTaskQueueSession, queueId)
		d.queueSessionMu.Unlock()
		return
	}

	t := queueSession.Front()
	d.executeTask(&t)
}

func (d *dispatch) executeTask(t *task) {
	select {
	case d.workerQueue <- *t:
		return
	default:
		// 使用原子操作读取 workerNum
		if atomic.LoadInt32(&d.workerNum) < d.maxConcurrentNum {
			var work worker
			work.start(&d.waitWorker, t, d)
			return
		}
	}

	d.workerQueue <- *t
}

func (d *dispatch) processTask(t *task) {
	atomic.StoreInt32(&d.idle, 0)

	//处理有排队任务
	if t.queueId != 0 {
		d.queueSessionMu.Lock()
		queueSession := d.mapTaskQueueSession[t.queueId]
		if queueSession == nil {
			queueSession = &queue.Deque[task]{}
			d.mapTaskQueueSession[t.queueId] = queueSession
		}
		queueLen := queueSession.Len()
		d.queueSessionMu.Unlock()

		//没有正在执行的任务，则直接执行
		if queueLen == 0 {
			d.executeTask(t)
		}

		d.queueSessionMu.Lock()
		queueSession.PushBack(*t)
		d.queueSessionMu.Unlock()
		return
	}

	//普通任务
	d.executeTask(t)
}

func (d *dispatch) processIdle() {
	// 改进：使用带超时的 channel 发送，避免无限阻塞
	// 如果 workerQueue 满了（所有 worker 都在执行任务），超时后不减少计数
	// worker 会在任务完成后自然退出，下次空闲检测时会再次尝试回收
	timeout := time.NewTimer(10 * time.Millisecond)
	defer timeout.Stop()

	select {
	case d.workerQueue <- task{}:
		// 成功发送退出信号，减少 worker 计数
		atomic.AddInt32(&d.workerNum, -1)
	case <-timeout.C:
		// 超时：说明所有 worker 都在执行任务，无法接收退出信号
		// 这种情况下，worker 会在任务完成后自然退出
		// 不减少计数，等待下次空闲检测时再次尝试
		currentNum := atomic.LoadInt32(&d.workerNum)
		minNum := atomic.LoadInt32(&d.minConcurrentNum)
		if currentNum > minNum {
			// 记录调试信息（可选）
			// log.Debug("worker queue is full, worker will exit after task completion")
		}
	}
}

func (d *dispatch) pushQueueTaskFinishEvent(queueId int64) {
	d.queueIdChannel <- queueId
}

func (d *dispatch) pushAsyncDoCallbackEvent(cb func(err error)) {
	if cb == nil {
		//不需要回调的情况
		return
	}

	d.cbChannel <- cb
}

func (d *dispatch) close() {
	atomic.StoreInt32(&d.minConcurrentNum, -1)
	d.cancel()

	// 等待所有 worker 退出
	d.waitWorker.Wait()

	// 处理剩余的回调，使用超时机制避免无限阻塞
	timeout := time.NewTimer(100 * time.Millisecond)
	defer timeout.Stop()

breakFor:
	for {
		select {
		case cb := <-d.cbChannel:
			if cb == nil {
				break breakFor
			}
			cb(nil)
			// 重置超时，因为还有回调在处理
			if !timeout.Stop() {
				<-timeout.C
			}
			timeout.Reset(100 * time.Millisecond)
		case <-timeout.C:
			// 超时后，如果所有 worker 都已退出且没有更多回调，直接退出
			// 使用非阻塞方式检查 channel 是否为空
			select {
			case cb := <-d.cbChannel:
				if cb == nil {
					break breakFor
				}
				cb(nil)
				timeout.Reset(100 * time.Millisecond)
			default:
				// channel 为空，可以退出
				break breakFor
			}
		}
	}

	d.waitDispatch.Wait()
}

func (d *dispatch) DoCallback(cb func(err error)) {
	defer func() {
		if r := recover(); r != nil {
			log.StackError(fmt.Sprint(r))
		}
	}()

	cb(nil)
}
