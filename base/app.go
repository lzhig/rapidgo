/**
 * @author [Bruce]
 * @email [lzhig@outlook.com]
 * @create date 2018-01-20 05:28:13
 * @modify date 2018-01-20 05:28:13
 * @desc [description]
 */

package base

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
	close(obj.exitChan)
}
