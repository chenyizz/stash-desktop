package app

import "github.com/wailsapp/wails/v3/pkg/application"

// wailsEmitter 把 Wails 的 application.App 适配成 logger.EventEmitter。
// 它是唯一连接 logger 和 Wails 的桥梁。
type wailsEmitter struct {
	app *application.App
}

func (e *wailsEmitter) Emit(name string, data ...any) {
	if e.app == nil {
		return
	}
	e.app.Event.Emit(name, data...)
}
