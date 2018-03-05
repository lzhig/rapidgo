/**
 * @author [Bruce]
 * @email [lzhig@outlook.com]
 * @create date 2018-01-21 01:40:51
 * @modify date 2018-01-21 01:40:51
 * @desc [description]
 */

package base

import (
	"context"
	"sync"
)

type GoroutineManager struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

// Init function
func (obj *GoroutineManager) Init() error {
	obj.ctx, obj.cancelFunc = context.WithCancel(context.Background())

	return nil
}

func (obj *GoroutineManager) Wait() {
	obj.wg.Wait()
}

func (obj *GoroutineManager) Exit() {
	obj.cancelFunc()
}

// CreateCancelContext 创建CancelContext, 方便goroutine管理
func (obj *GoroutineManager) CreateCancelContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(obj.ctx)
}

// GoRoutine 管理goroutine的创建及退出
// 例如:
// ctx, _ := obj.app.CreateCancelContext()
// obj.app.GoRoutine(ctx,
// 	func(ctx context.Context, args ...interface{}) {
// 		fmt.Println(args)
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				fmt.Println("goroutine 1 exit.")
// 				return
// 			}
// 		}
// 	}, 1)
func (obj *GoroutineManager) GoRoutine(ctx context.Context, f func(context.Context)) {
	obj.wg.Add(1)
	go func() {
		defer obj.wg.Done()
		f(ctx)
	}()
}

func (obj *GoroutineManager) GoRoutineArgs(ctx context.Context, f func(context.Context, ...interface{}), args ...interface{}) {
	obj.wg.Add(1)
	go func() {
		defer obj.wg.Done()
		f(ctx, args...)
	}()
}
