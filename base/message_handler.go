package base

// MessageHandlerFunc type
type MessageHandlerFunc func(interface{})

type MessageHandler interface {
	Init()
	Close()
	AddMessageHandler(interface{}, MessageHandlerFunc)
	AddBusyHandler(interface{}, MessageHandlerFunc)
	Handle(interface{}, interface{})
	HandleNoHandler(interface{}, interface{})
	HandleNoBusyHandler(interface{}, interface{})
}

type MessageHandlerImpl struct {
	messageDispatcher map[interface{}]MessageHandlerFunc
}

func (obj *MessageHandlerImpl) Init() {
	obj.messageDispatcher = make(map[interface{}]MessageHandlerFunc)
}

func (obj *MessageHandlerImpl) SetMessageHandler(k interface{}, f MessageHandlerFunc) {
	obj.messageDispatcher[k] = f
}

func (obj *MessageHandlerImpl) Handle(k interface{}, v interface{}) bool {
	if handler, ok := obj.messageDispatcher[k]; ok {
		handler(v)
		return true
	}
	return false
}
