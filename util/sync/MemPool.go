package sync

import (
	sysSync "sync"
)

type Pool struct {
	C        chan interface{} //最大缓存的数量
	syncPool sysSync.Pool
}

func (pool *Pool) Get() interface{} {
	select {
	case d := <-pool.C:
		return d
	default:
		return pool.syncPool.Get()
	}
}

func (pool *Pool) Put(data interface{}) {
	select {
	case pool.C <- data:
	default:
		pool.syncPool.Put(data)
	}

}

func NewPool(C chan interface{}, New func() interface{}) *Pool {
	var p Pool
	p.C = C
	p.syncPool.New = New
	return &p
}
