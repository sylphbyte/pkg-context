package context

import (
	"sync"
)

// event 独立事件系统
type event struct {
	subscribers sync.Map   // map[string][]EventHandler
	mu          sync.Mutex // 用于保护订阅操作
}

// EventHandler 定义事件处理函数
type EventHandler func(ctx Context, payload interface{})

// NewEvent 创建新的事件系统
func newEvent() *event {
	return &event{
		subscribers: sync.Map{},
	}
}

// On 订阅事件
func (es *event) On(eventName string, handlers ...EventHandler) {
	es.mu.Lock()
	defer es.mu.Unlock()

	stored, _ := es.subscribers.LoadOrStore(eventName, []EventHandler{})
	if storedHandlers, ok := stored.([]EventHandler); ok {
		updated := append(storedHandlers, handlers...)
		es.subscribers.Store(eventName, updated)
	} else {
		// 处理意外类型（理论上不应发生）
		es.subscribers.Store(eventName, handlers)
	}
}

// Off 取消订阅特定事件的某个处理函数
func (es *event) Off(eventName string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	if _, ok := es.subscribers.Load(eventName); ok {
		es.subscribers.Delete(eventName)
	}
}

// Emit 触发事件
func (es *event) Emit(ctx Context, eventName string, payload interface{}) {
	if handlers, ok := es.subscribers.Load(eventName); ok {
		if storedHandlers, ok := handlers.([]EventHandler); ok {
			for _, h := range storedHandlers {
				h(ctx, payload)
			}
			// wg.Wait()
		}
	}
}

// AsyncEmit 异步触发事件
func (es *event) AsyncEmit(ctx Context, eventName string, payload interface{}) {
	if handlers, ok := es.subscribers.Load(eventName); ok {
		if storedHandlers, ok := handlers.([]EventHandler); ok {
			var wg sync.WaitGroup
			for _, h := range storedHandlers {
				wg.Add(1)

				// 使用SafeGo启动goroutine
				handler := h // 创建副本避免闭包问题
				SafeGo(ctx, "x.event.AsyncEmit", func() {
					defer wg.Done()
					handler(ctx, payload)
				})
			}

			// 等待所有事件处理完成
			wg.Wait()
		}
	}
}

// AsyncEmitNoWait 异步触发事件但不等待完成
func (es *event) AsyncEmitNoWait(ctx Context, eventName string, payload interface{}) {
	if handlers, ok := es.subscribers.Load(eventName); ok {
		if storedHandlers, ok := handlers.([]EventHandler); ok {
			for _, h := range storedHandlers {
				// 使用SafeGo启动goroutine
				handler := h // 创建副本避免闭包问题
				SafeGo(ctx, "x.event.AsyncEmitNoWait", func() {
					handler(ctx, payload)
				})
			}
		}
	}
}
