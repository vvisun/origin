package event

import "github.com/duanhf2012/origin/v2/util/usync"

// eventPool的内存池,缓存Event
const defaultMaxEventChannelNum = 2000000

var eventPool = usync.NewPoolEx(make(chan usync.IPoolData, defaultMaxEventChannelNum), func() usync.IPoolData {
	return &Event{}
})

func NewEvent() *Event {
	return eventPool.Get().(*Event)
}

func DeleteEvent(event IEvent) {
	eventPool.Put(event.(usync.IPoolData))
}

func SetEventPoolSize(eventPoolSize int) {
	eventPool = usync.NewPoolEx(make(chan usync.IPoolData, eventPoolSize), func() usync.IPoolData {
		return &Event{}
	})
}
