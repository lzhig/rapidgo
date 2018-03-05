/**
 * @author [Bruce]
 * @email [lzhig@outlook.com]
 * @create date 2018-01-20 05:28:13
 * @modify date 2018-01-20 05:28:13
 * @desc [description]
 */

package base

import "context"

// App type
type App struct {
	goroutineManager GoroutineManager

	exitChan chan struct{}
}

// Init function
func (obj *App) Init() error {
	obj.exitChan = make(chan struct{})
	obj.goroutineManager.Init()

	return nil
}

// Start function
func (obj *App) Start() {
	select {
	case <-obj.exitChan:
		break
	}
	obj.goroutineManager.Wait()
}

// Exit function
func (obj *App) Exit() {
	obj.goroutineManager.Exit()
	obj.exitChan <- struct{}{}
}

// CreateCancelContext 创建CancelContext, 方便goroutine管理
func (obj *App) CreateCancelContext() (context.Context, context.CancelFunc) {
	return obj.goroutineManager.CreateCancelContext()
}

// GoRoutine function
func (obj *App) GoRoutine(ctx context.Context, f func(context.Context)) {
	obj.goroutineManager.GoRoutine(ctx, f)
}

// GoRoutine function
func (obj *App) GoRoutineArgs(ctx context.Context, f func(context.Context, ...interface{}), args ...interface{}) {
	obj.goroutineManager.GoRoutineArgs(ctx, f, args...)
}
