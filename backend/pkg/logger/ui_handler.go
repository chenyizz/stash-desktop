package logger

import (
	"context"
	"log/slog"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// UIHandler 把日志推送到 Wails 前端
type UIHandler struct {
	app      *application.App
	minLevel slog.Level
	attrs    []slog.Attr
	group    string
}

func NewUIHandler(app *application.App, minLevel slog.Level) *UIHandler {
	return &UIHandler{app: app, minLevel: minLevel}
}

func (h *UIHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func (h *UIHandler) Handle(ctx context.Context, r slog.Record) error {
	// 在原来基础上，把 h.attrs 和 h.group 也带上
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

	h.app.Event.Emit("app:log", map[string]any{
		"level":   r.Level.String(),
		"message": r.Message,
		"time":    r.Time.Format("15:04:05"),
		"attrs":   attrs,
	})
	return nil
}

func (h *UIHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &UIHandler{app: h.app, minLevel: h.minLevel, attrs: attrs, group: h.group}
}

func (h *UIHandler) WithGroup(name string) slog.Handler {
	return &UIHandler{app: h.app, minLevel: h.minLevel, attrs: h.attrs, group: name}
}
