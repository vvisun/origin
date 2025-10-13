package usync

import (
	"fmt"
	sysSync "sync"
	"sync/atomic"
)

// RefPoolObject 引用池对象接口
type RefPoolObject interface {
	OnDtor()
	getPoolRef() *int64
}

// BasePoolObject 基础池对象，包含引用计数
type BasePoolObject struct {
	poolRef int64 //原子操作的池对象引用标记。为0表示未被池管理，大于0表示在池中，防止重复放入，导致从池中取对象时，取到重复的对象
}

// getPoolRef 获取池引用计数
func (bpo *BasePoolObject) getPoolRef() *int64 {
	return &bpo.poolRef
}

// OnDtor 析构函数
func (bpo *BasePoolObject) OnDtor() {
	// 默认实现为空，子类可以重写
}

// RefPool 引用计数对象池
type RefPool[T RefPoolObject] struct {
	pool    sysSync.Pool
	size    int64 // 使用原子操作
	newFunc func() T
}

// NewRefPool 创建引用计数对象池
func NewRefPool[T RefPoolObject](newFunc func() T) *RefPool[T] {
	return &RefPool[T]{
		pool: sysSync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
		newFunc: newFunc,
		size:    0,
	}
}

// Get 获取对象
func (rp *RefPool[T]) Get() T {
	obj := rp.pool.Get().(T)

	if IsNil(obj) || !IsPointer(obj) || IsDoublePointer(obj) {
		panic("obj is nil or not ptr")
	}

	// 原子操作减少size
	if atomic.LoadInt64(&rp.size) > 0 {
		atomic.AddInt64(&rp.size, -1)
	}

	// 重置引用计数为0，表示对象已被取出
	atomic.StoreInt64(obj.getPoolRef(), 0)

	return obj
}

// Put 归还对象
func (rp *RefPool[T]) Put(obj T) {
	// 检查对象是否为nil（对于指针类型）
	if IsNil(obj) || !IsPointer(obj) || IsDoublePointer(obj) {
		fmt.Println("obj is nil or not ptr")
		return
	}

	// 检查引用计数，防止重复放入
	poolRef := obj.getPoolRef()
	if atomic.LoadInt64(poolRef) > 0 {
		fmt.Println("object already in pool, ignoring put")
		return
	}

	// 设置引用计数为1，表示对象在池中
	atomic.StoreInt64(poolRef, 1)

	obj.OnDtor()
	rp.pool.Put(obj)
	atomic.AddInt64(&rp.size, 1)
}

// Clear 清空池
func (rp *RefPool[T]) Clear() {
	newFunc := rp.newFunc
	rp.pool = sysSync.Pool{
		New: func() interface{} {
			return newFunc()
		},
	}
	atomic.StoreInt64(&rp.size, 0)
}

// Size 获取池大小（原子操作）
func (rp *RefPool[T]) Size() int64 {
	return atomic.LoadInt64(&rp.size)
}
