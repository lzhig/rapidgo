package base

import (
	"container/list"
	"fmt"
)

///////////////////////////////////////////////////////////////////////////////
// ErrorID 错误码
type ErrorID int

const (
	// ErrorNone 正确
	ErrorNone = ErrorID(iota)

	// ErrorInvalidStateID 非法的状态ID
	ErrorInvalidStateID

	//ErrorNotExistStateID 不存在的状态ID
	ErrorNotExistStateID

	//ErrorInternalError 内部错误
	ErrorInternalError
)

// StateError 状态错误体
type StateError struct {
	errID     ErrorID
	errString string
}

func (e StateError) Error() string {
	return e.errString
}

///////////////////////////////////////////////////////////////////////////////
// StateID 状态ID
type StateID int

// StateInvalidID 非法的状态ID，用于状态机初始值
const StateInvalidID StateID = StateID(0)

///////////////////////////////////////////////////////////////////////////////
type commandID uint

const (
	commandPUSH = commandID(iota)
	commandCHANGE
	commandPOP
	commandPOPALL
)

type command struct {
	commandType commandID
	stateID     StateID
}

///////////////////////////////////////////////////////////////////////////////
// StateListener 状态监听
type StateListener interface {
	stateChange(newStateID StateID, oldStateID StateID)
}

// StateInterface 状态接口
type StateInterface interface {
	GetStateID() StateID           // 返回当前状态ID
	OnEnter(prevStateID StateID)   // 从前一个状态切换到此状态
	OnExit(nextStateID StateID)    // 从此状态切换到下一状态
	OnSuspend(nextStateID StateID) // 临时挂起此状态
	OnResume(prevStateID StateID)  // 恢复此状态
	Update()                       // 更新函数
	OnMessage(data []byte)         // 处理收到的消息
}

///////////////////////////////////////////////////////////////////////////////
// StateManager 状态管理类
type StateManager struct {
	listeners list.List

	states map[StateID]StateInterface // 注册的所有状态
	//currentState StateInterface             // 当前状态

	commandQueue list.List // 命令队列
	currentState list.List // 当前状态队列
}

// Initialize 初始化
func (sm *StateManager) Initialize() {
	sm.states = make(map[StateID]StateInterface)
}

// Terminate 中断
func (sm *StateManager) Terminate() {

}

// AddListener 添加监听
func (sm *StateManager) AddListener(listener StateListener) {
	sm.listeners.PushBack(listener)
}

// RemoveListener 移除监听
func (sm *StateManager) RemoveListener(listener StateListener) {
	for e := sm.listeners.Front(); e != nil; e = e.Next() {
		if e.Value == listener {
			sm.listeners.Remove(e)
			return
		}
	}
}

// RegisterState 注册状态
func (sm *StateManager) RegisterState(stateID StateID, state StateInterface) *StateError {
	if stateID == StateInvalidID {
		return &StateError{ErrorInvalidStateID, "Invalid State ID"}
	}
	sm.states[stateID] = state

	return nil
}

// GetState 获取状态
func (sm *StateManager) GetState(stateID StateID) StateInterface {
	return sm.states[stateID]
}

func (sm *StateManager) GetCurrentState() StateInterface {
	if sm.currentState.Len() == 0 {
		return nil
	}

	state := sm.currentState.Front().Value.(StateInterface)
	return state
}

func (sm *StateManager) SendMessage(data []byte) {
	state := sm.GetCurrentState()
	if state != nil {
		state.OnMessage(data)
	}
}

// ChangeState 改变状态
func (sm *StateManager) ChangeState(stateID StateID, flushCommandQueue bool) {
	if flushCommandQueue {
		sm.commandQueue.Init()
	}
	sm.commandQueue.PushFront(&command{commandType: commandCHANGE, stateID: stateID})

	//fmt.Println("ChangeState command type:", commandCHANGE, ", state id:", stateID)
}

func (sm *StateManager) ChangeStateImmediate(stateID StateID) {
	sm.ChangeState(stateID, false)
	sm.updateCommandQueue()
}

func (sm *StateManager) PushState(stateID StateID, flushCommandQueue bool) {
	if flushCommandQueue {
		sm.commandQueue.Init()
	}
	sm.commandQueue.PushFront(&command{commandType: commandPUSH, stateID: stateID})
}

func (sm *StateManager) updateCommandQueue() {
	for sm.commandQueue.Len() > 0 {
		elm := sm.commandQueue.Front()
		cmd, ok := elm.Value.(*command)

		//fmt.Println("cmd type:", cmd.commandType, ", state id:", cmd.stateID)
		sm.commandQueue.Remove(elm)

		if !ok {
			panic(StateError{ErrorInternalError, "Invalid type of command."})
		}

		if cmd.commandType == commandPUSH {
			prevState := sm.GetCurrentState()
			prevStateID := StateInvalidID
			if prevState != nil {
				prevState.OnSuspend(cmd.stateID)
				prevStateID = prevState.GetStateID()
			}
			currentState := sm.states[cmd.stateID]
			sm.currentState.PushFront(currentState)
			currentState.OnEnter(prevStateID)
		} else if cmd.commandType == commandPOP {
			prevState := sm.GetCurrentState()
			prevStateID := StateInvalidID
			if prevState != nil {
				currentStateID := StateInvalidID
				if sm.currentState.Len() >= 2 {
					currentStateID = sm.currentState.Front().Next().Value.(StateInterface).GetStateID()
				}
				prevState.OnExit(currentStateID)

				sm.currentState.Remove(sm.currentState.Front())

				currentState := sm.GetCurrentState()
				if currentState != nil {
					currentState.OnResume(prevStateID)
				}
			} else {
				panic(&StateError{ErrorInternalError, "Cannot pop state because no state is in queue."})
			}
		} else if cmd.commandType == commandPOPALL {
			for sm.currentState.Len() > 1 {
				prevState := sm.GetCurrentState()
				prevStateID := StateInvalidID
				currentStateID := StateInvalidID

				currentState := sm.currentState.Front().Next().Value.(StateInterface)
				currentStateID = currentState.GetStateID()
				prevState.OnExit(currentStateID)

				sm.currentState.Remove(sm.currentState.Front())

				currentState.OnResume(prevStateID)
			}
		} else if cmd.commandType == commandCHANGE {
			prevState := sm.GetCurrentState()
			if prevState.GetStateID() != cmd.stateID {
				prevState.OnExit(cmd.stateID)
				sm.currentState.Remove(sm.currentState.Front())
				sm.currentState.PushFront(sm.states[cmd.stateID])
				sm.GetCurrentState().OnEnter(prevState.GetStateID())
			} else {
				panic(&StateError{ErrorInternalError, fmt.Sprintf("Cannot change to same state: %d", cmd.stateID)})
			}
		}

	}
}

// Update 更新状态
func (sm *StateManager) Update() {
	sm.updateCommandQueue()

	currentState := sm.GetCurrentState()
	if currentState != nil {
		currentState.Update()
	}
}
