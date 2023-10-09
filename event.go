package goearth

type VoidEvent struct {
	handlers []VoidHandler
}
type VoidHandler func()

func (e *VoidEvent) Register(handler VoidHandler) {
	e.handlers = append(e.handlers, handler)
}

func (e *VoidEvent) Dispatch() {
	for _, handler := range e.handlers {
		handler()
	}
}

type Event[T any] struct {
	handlers []EventHandler[T]
}
type EventHandler[T any] func(e *T)

func (e *Event[T]) Register(handler EventHandler[T]) {
	e.handlers = append(e.handlers, handler)
}

func (e *Event[T]) Dispatch(args *T) {
	for _, handler := range e.handlers {
		handler(args)
	}
}

type InitArgs struct {
	Connected bool
}

type InitEvent = Event[InitArgs]
type InitHandler = EventHandler[InitArgs]

type ConnectArgs struct {
	Host     string
	Port     int
	Client   Client
	Messages []MsgInfo
}

type ConnectEvent = Event[ConnectArgs]
type ConnectHandler = EventHandler[ConnectArgs]

type InterceptEvent = Event[InterceptArgs]
type InterceptHandler = EventHandler[InterceptArgs]
