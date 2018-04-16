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
	closeChan         chan struct{}
	waitChan          chan struct{}
	protoChan         chan [2]interface{}
	messageDispatcher map[interface{}]MessageHandlerFunc
	busyDispatcher    map[interface{}]MessageHandlerFunc
}

func (obj *MessageHandlerImpl) Init(queueLength int) {
	obj.closeChan = make(chan struct{})
	obj.waitChan = make(chan struct{}, 1)
	obj.protoChan = make(chan [2]interface{}, queueLength)
	obj.messageDispatcher = make(map[interface{}]MessageHandlerFunc)
	obj.busyDispatcher = make(map[interface{}]MessageHandlerFunc)

	go func() {
		defer LogPanic()
		obj.loop()
	}()
}

func (obj *MessageHandlerImpl) Close(wait bool) {
	obj.closeChan <- struct{}{}
	if wait {
		<-obj.waitChan
	}
	close(obj.waitChan)
}

func (obj *MessageHandlerImpl) loop() {
	defer func() { obj.waitChan <- struct{}{} }()
	for {
		select {
		case <-obj.closeChan:
			return

		case p := <-obj.protoChan:
			if f, ok := obj.messageDispatcher[p[0]]; ok {
				f(p[1])
			} else {
				obj.HandleNoHandler(p[0], p[1])
			}
		}
	}
}

func (obj *MessageHandlerImpl) AddMessageHandler(k interface{}, f MessageHandlerFunc) {
	obj.messageDispatcher[k] = f
}

func (obj *MessageHandlerImpl) AddBusyHandler(k interface{}, f MessageHandlerFunc) {
	obj.busyDispatcher[k] = f
}

func (obj *MessageHandlerImpl) Handle(k interface{}, v interface{}) {
	select {
	case obj.protoChan <- [2]interface{}{k, v}:
	default:
		obj.handleBusy(k, v)
	}
}

func (obj *MessageHandlerImpl) handleBusy(k interface{}, v interface{}) {
	if f, ok := obj.busyDispatcher[k]; ok {
		f(v)
	} else {
		obj.HandleNoBusyHandler(k, v)
	}
}

func (obj *MessageHandlerImpl) HandleNoHandler(interface{}, interface{}) {

}

func (obj *MessageHandlerImpl) HandleNoBusyHandler(interface{}, interface{}) {

}
