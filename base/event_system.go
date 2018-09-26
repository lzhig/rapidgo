package base

// EventID 事件ID
type EventID int

// EventHandler 事件处理器
type EventHandler func([]interface{})

type event struct {
	id   EventID
	data []interface{}
}

// EventSystem 事件处理系统
type EventSystem struct {
	closeChan   chan struct{}
	waitChan    chan struct{}
	eventsQueue chan *event
	handlers    map[EventID]EventHandler
	sendFunc    func(id EventID, data []interface{}) bool
}

// Init 初始化
func (obj *EventSystem) Init(queueLength int, blocked bool) {
	obj.closeChan = make(chan struct{}, 1)
	obj.waitChan = make(chan struct{}, 1)
	obj.eventsQueue = make(chan *event, queueLength)
	obj.handlers = make(map[EventID]EventHandler)
	if blocked {
		obj.sendFunc = obj.sendBlock
	} else {
		obj.sendFunc = obj.sendNoBlock
	}

	go func() {
		defer LogPanic()
		obj.loop()
	}()
}

// Close 关闭
func (obj *EventSystem) Close(wait bool) {
	obj.closeChan <- struct{}{}
	if wait {
		<-obj.waitChan
	}
}

// SetEventHandler 设置事件处理器
func (obj *EventSystem) SetEventHandler(id EventID, f EventHandler) {
	obj.handlers[id] = f
}

// Send 发送事件
func (obj *EventSystem) Send(id EventID, data []interface{}) bool {
	return obj.sendFunc(id, data)
}

func (obj *EventSystem) sendBlock(id EventID, data []interface{}) bool {
	obj.eventsQueue <- &event{id: id, data: data}
	return true
}

func (obj *EventSystem) sendNoBlock(id EventID, data []interface{}) bool {
	select {
	case obj.eventsQueue <- &event{id: id, data: data}:
		return true
	default:
		return false
	}
}

func (obj *EventSystem) loop() {
	//defer LogInfo("exit loop")
	defer func() { obj.waitChan <- struct{}{} }()
	for {
		select {
		case <-obj.closeChan:
			return
		case e := <-obj.eventsQueue:
			if handler, ok := obj.handlers[e.id]; ok {
				//LogInfo(e.id)
				handler(e.data)
			} else {
				LogError("No handler for the event. id:", e.id)
			}
		}
	}
}
