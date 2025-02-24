package service

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/duanhf2012/origin/v2/concurrent"
	"github.com/duanhf2012/origin/v2/event"
	"github.com/duanhf2012/origin/v2/log"
	rpcHandle "github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/util/timer"
	"slices"
)

const InitModuleId = 1e9

type IModule interface {
	concurrent.IConcurrent
	SetModuleId(moduleId uint32) bool
	GetModuleId() uint32
	AddModule(module IModule) (uint32, error)
	GetModule(moduleId uint32) IModule
	GetAncestor() IModule
	ReleaseModule(moduleId uint32)
	NewModuleId() uint32
	GetParent() IModule
	OnInit() error
	OnRelease()
	getBaseModule() IModule
	GetService() IService
	GetModuleName() string
	GetEventProcessor() event.IEventProcessor
	NotifyEvent(ev event.IEvent)
}

type IModuleTimer interface {
	SafeAfterFunc(d time.Duration, cb func(*timer.Timer)) *timer.Timer
	SafeCronFunc(cronExpr *timer.CronExpr, cb func(*timer.Cron)) *timer.Cron
	SafeNewTicker(d time.Duration, cb func(*timer.Ticker)) *timer.Ticker
}

type Module struct {
	rpcHandle.IRpcHandler
	moduleId         uint32             //模块Id
	moduleName       string             //模块名称
	parent           IModule            //父亲
	self             IModule            //自己
	child            []IModule //孩子们
	mapActiveTimer   map[timer.ITimer]struct{}
	mapActiveIdTimer map[uint64]timer.ITimer
	dispatcher       *timer.Dispatcher //timer

	//根结点
	ancestor     IModule            //始祖
	seedModuleId uint32             //模块id种子
	descendants  map[uint32]IModule //始祖的后裔们

	//事件管道
	eventHandler event.IEventHandler
	concurrent.IConcurrent
}

func (m *Module) SetModuleId(moduleId uint32) bool {
	if m.moduleId > 0 {
		return false
	}

	m.moduleId = moduleId
	return true
}

func (m *Module) GetModuleId() uint32 {
	return m.moduleId
}

func (m *Module) GetModuleName() string {
	return m.moduleName
}

func (m *Module) OnInit() error {
	return nil
}

func (m *Module) AddModule(module IModule) (uint32, error) {
	//没有事件处理器不允许加入其他模块
	if m.GetEventProcessor() == nil {
		return 0, fmt.Errorf("module %+v Event Processor is nil", m.self)
	}

	pAddModule := module.getBaseModule().(*Module)
	if pAddModule.GetModuleId() == 0 {
		pAddModule.moduleId = m.NewModuleId()
	}

	_,ok := m.ancestor.getBaseModule().(*Module).descendants[module.GetModuleId()]
	if ok == true {
		return 0, fmt.Errorf("exists module id %d", module.GetModuleId())
	}
	pAddModule.IRpcHandler = m.IRpcHandler
	pAddModule.self = module
	pAddModule.parent = m.self
	pAddModule.dispatcher = m.GetAncestor().getBaseModule().(*Module).dispatcher
	pAddModule.ancestor = m.ancestor
	pAddModule.moduleName = reflect.Indirect(reflect.ValueOf(module)).Type().Name()
	pAddModule.eventHandler = event.NewEventHandler()
	pAddModule.eventHandler.Init(m.eventHandler.GetEventProcessor())
	pAddModule.IConcurrent = m.IConcurrent

	m.child = append(m.child,module)
	m.ancestor.getBaseModule().(*Module).descendants[module.GetModuleId()] = module

	err := module.OnInit()
	if err != nil {
		delete(m.ancestor.getBaseModule().(*Module).descendants, module.GetModuleId())
		m.child = m.child[:len(m.child)-1]
		log.Error("module OnInit error",log.String("ModuleName",module.GetModuleName()),log.ErrorField("err",err))
		return 0, err
	}

	log.Debug("Add module " + module.GetModuleName() + " completed")
	return module.GetModuleId(), nil
}

func (m *Module) ReleaseModule(moduleId uint32) {
	pModule := m.GetModule(moduleId).getBaseModule().(*Module)
	pModule.self.OnRelease()
	log.Debug("Release module " + pModule.GetModuleName())

	for i:=len(pModule.child)-1; i>=0; i-- {
		m.ReleaseModule(pModule.child[i].GetModuleId())
	}

	pModule.GetEventHandler().Destroy()

	for pTimer := range pModule.mapActiveTimer {
		pTimer.Cancel()
	}

	for _, t := range pModule.mapActiveIdTimer {
		t.Cancel()
	}

	m.child = slices.DeleteFunc(m.child, func(module IModule) bool {
		return module.GetModuleId() == moduleId
	})

	delete(m.ancestor.getBaseModule().(*Module).descendants, moduleId)

	//清理被删除的Module
	pModule.self = nil
	pModule.parent = nil
	pModule.child = nil
	pModule.mapActiveTimer = nil
	pModule.dispatcher = nil
	pModule.ancestor = nil
	pModule.descendants = nil
	pModule.IRpcHandler = nil
	pModule.mapActiveIdTimer = nil
}

func (m *Module) NewModuleId() uint32 {
	m.ancestor.getBaseModule().(*Module).seedModuleId += 1
	return m.ancestor.getBaseModule().(*Module).seedModuleId
}

var timerSeedId uint32

func (m *Module) GenTimerId() uint64 {
	for {
		newTimerId := (uint64(m.GetModuleId()) << 32) | uint64(atomic.AddUint32(&timerSeedId, 1))
		if _, ok := m.mapActiveIdTimer[newTimerId]; ok == true {
			continue
		}

		return newTimerId
	}
}

func (m *Module) GetAncestor() IModule {
	return m.ancestor
}

func (m *Module) GetModule(moduleId uint32) IModule {
	iModule, ok := m.GetAncestor().getBaseModule().(*Module).descendants[moduleId]
	if ok == false {
		return nil
	}
	return iModule
}

func (m *Module) getBaseModule() IModule {
	return m
}

func (m *Module) GetParent() IModule {
	return m.parent
}

func (m *Module) OnCloseTimer(t timer.ITimer) {
	delete(m.mapActiveIdTimer, t.GetId())
	delete(m.mapActiveTimer, t)
}

func (m *Module) OnAddTimer(t timer.ITimer) {
	if t != nil {
		if m.mapActiveTimer == nil {
			m.mapActiveTimer = map[timer.ITimer]struct{}{}
		}

		m.mapActiveTimer[t] = struct{}{}
	}
}

// Deprecated: this function simply calls SafeAfterFunc
func (m *Module) AfterFunc(d time.Duration, cb func(*timer.Timer)) *timer.Timer {
	if m.mapActiveTimer == nil {
		m.mapActiveTimer = map[timer.ITimer]struct{}{}
	}

	return m.dispatcher.AfterFunc(d, nil, cb, m.OnCloseTimer, m.OnAddTimer)
}

// Deprecated: this function simply calls SafeCronFunc
func (m *Module) CronFunc(cronExpr *timer.CronExpr, cb func(*timer.Cron)) *timer.Cron {
	if m.mapActiveTimer == nil {
		m.mapActiveTimer = map[timer.ITimer]struct{}{}
	}

	return m.dispatcher.CronFunc(cronExpr, nil, cb, m.OnCloseTimer, m.OnAddTimer)
}

// Deprecated: this function simply calls SafeNewTicker
func (m *Module) NewTicker(d time.Duration, cb func(*timer.Ticker)) *timer.Ticker {
	if m.mapActiveTimer == nil {
		m.mapActiveTimer = map[timer.ITimer]struct{}{}
	}

	return m.dispatcher.TickerFunc(d, nil, cb, m.OnCloseTimer, m.OnAddTimer)
}

func (m *Module) cb(*timer.Timer) {

}

func (m *Module) SafeAfterFunc(timerId *uint64, d time.Duration, AdditionData interface{}, cb func(uint64, interface{})) {
	if m.mapActiveIdTimer == nil {
		m.mapActiveIdTimer = map[uint64]timer.ITimer{}
	}

	if *timerId != 0 {
		m.CancelTimerId(timerId)
	}

	*timerId = m.GenTimerId()
	t := m.dispatcher.AfterFunc(d, cb, nil, m.OnCloseTimer, m.OnAddTimer)
	t.AdditionData = AdditionData
	t.Id = *timerId
	m.mapActiveIdTimer[*timerId] = t
}

func (m *Module) SafeCronFunc(cronId *uint64, cronExpr *timer.CronExpr, AdditionData interface{}, cb func(uint64, interface{})) {
	if m.mapActiveIdTimer == nil {
		m.mapActiveIdTimer = map[uint64]timer.ITimer{}
	}

	*cronId = m.GenTimerId()
	c := m.dispatcher.CronFunc(cronExpr, cb, nil, m.OnCloseTimer, m.OnAddTimer)
	c.AdditionData = AdditionData
	c.Id = *cronId
	m.mapActiveIdTimer[*cronId] = c
}

func (m *Module) SafeNewTicker(tickerId *uint64, d time.Duration, AdditionData interface{}, cb func(uint64, interface{})) {
	if m.mapActiveIdTimer == nil {
		m.mapActiveIdTimer = map[uint64]timer.ITimer{}
	}

	*tickerId = m.GenTimerId()
	t := m.dispatcher.TickerFunc(d, cb, nil, m.OnCloseTimer, m.OnAddTimer)
	t.AdditionData = AdditionData
	t.Id = *tickerId
	m.mapActiveIdTimer[*tickerId] = t
}

func (m *Module) CancelTimerId(timerId *uint64) bool {
	if timerId == nil || *timerId == 0 {
		log.Warn("timerId is invalid")
		return false
	}

	if m.mapActiveIdTimer == nil {
		log.Error("mapActiveIdTimer is nil")
		return false
	}

	t, ok := m.mapActiveIdTimer[*timerId]
	if ok == false {
		log.StackError("cannot find timer id ", log.Uint64("timerId", *timerId))
		return false
	}

	t.Cancel()
	*timerId = 0
	return true
}

func (m *Module) OnRelease() {
}

func (m *Module) GetService() IService {
	return m.GetAncestor().(IService)
}

func (m *Module) GetEventProcessor() event.IEventProcessor {
	return m.eventHandler.GetEventProcessor()
}

func (m *Module) NotifyEvent(ev event.IEvent) {
	m.eventHandler.NotifyEvent(ev)
}

func (m *Module) GetEventHandler() event.IEventHandler {
	return m.eventHandler
}
