package app

import (
	"context"
	"time"

	"case/backend/pkg/logger"
)

// EventScanComplete 在扫描/清理任务完成时由后端推送给前端。
// 契约冻结：仅此事件 + payload {at:int64}（不新增 started/failed/jobType）。
const EventScanComplete = "scan:complete"

// scanSubscriber 只暴露 watcher 所需的最小能力，便于测试与解耦。
type scanSubscriber interface {
	ScanSubscribe(ctx context.Context) <-chan bool
}

// watchScanEvents 把扫描完成信号转换为前端事件；ctx 取消或 channel 关闭即退出。
func watchScanEvents(ctx context.Context, sub scanSubscriber, emit func(name string, data ...any)) {
	ch := sub.ScanSubscribe(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-ch:
			if !ok {
				return
			}

			emit(EventScanComplete, map[string]any{"at": time.Now().Unix()})
		}
	}
}

// wailsEmitter 满足 logger 的事件接口，这里断言其仍可直接用于推送应用事件。
var _ logger.EventEmitter = (*wailsEmitter)(nil)
