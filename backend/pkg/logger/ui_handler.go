package logger

import (
	"context"
	"log/slog"
)

// EventEmitter 是 UIHandler 对外的唯一依赖。
// 任何能向 UI 发送事件的对象都可以实现这个接口。
// logger 包本身不依赖任何 UI 框架。
type EventEmitter interface {
	Emit(name string, data ...any)
}

// UIHandler 把日志推送到 UI。
type UIHandler struct {
	emitter  EventEmitter
	minLevel slog.Level
	attrs    []slog.Attr
	group    string
}

// NewUIHandler 创建一个 UIHandler。
// emitter 为 nil 时，Handle 会静默忽略。
func NewUIHandler(emitter EventEmitter, minLevel slog.Level) *UIHandler {
	return &UIHandler{emitter: emitter, minLevel: minLevel}
}

func (h *UIHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func (h *UIHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.emitter == nil {
		return nil
	}

	attrs := make(map[string]any)
	for _, a := range h.attrs {
		attrs[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		key := a.Key
		if h.group != "" {
			key = h.group + "." + key
		}
		attrs[key] = a.Value.Any()
		return true
	})

	h.emitter.Emit("app:log", map[string]any{
		"level":   r.Level.String(),
		"message": r.Message,
		"time":    r.Time.Format("15:04:05"),
		"attrs":   attrs,
	})
	return nil
}

func (h *UIHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &UIHandler{emitter: h.emitter, minLevel: h.minLevel, attrs: attrs, group: h.group}
}

func (h *UIHandler) WithGroup(name string) slog.Handler {
	return &UIHandler{emitter: h.emitter, minLevel: h.minLevel, attrs: h.attrs, group: name}
}