package sync

import (
	sysSync "sync"
)

type IPoolData interface {
	Reset()
	IsRef() bool
	Ref()
	UnRef()
}

type PoolEx struct {
	C        chan IPoolData //最大缓存的数量
	syncPool sysSync.Pool
}

func (pool *PoolEx) Get() IPoolData {
	select {
	case d := <-pool.C:
		if d.IsRef() {
			panic("Pool data is in use.")
		}

		d.Ref()
		return d
	default:
		data := pool.syncPool.Get().(IPoolData)
		if data.IsRef() {
			panic("Pool data is in use.")
		}

		data.Ref()
		return data
	}
}

func (pool *PoolEx) Put(data IPoolData) {
	if !data.IsRef() {
		panic("Repeatedly freeing memory")
	}
	//提前解引用，防止递归释放
	data.UnRef()
	data.Reset()
	//再次解引用，防止Rest时错误标记
	data.UnRef()
	select {
	case pool.C <- data:
	default:
		pool.syncPool.Put(data)
	}
}

func NewPoolEx(C chan IPoolData, New func() IPoolData) *PoolEx {
	var pool PoolEx
	pool.C = C
	pool.syncPool.New = func() interface{} {
		return New()
	}
	return &pool
}
