package app

import (
	"net/http"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	coverURLPrefix          = "/covers/"
	attachmentURLPrefix     = "/attachments/"
	performerImageURLPrefix = "/performers/"
)

// AssetMiddleware 仅按前缀分发到对应 handler，保持中间件无业务分支。
// 它是包级函数而非 App 方法，避免被 Wails 绑定为前端可调方法。
func AssetMiddleware(a *App) application.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasPrefix(r.URL.Path, coverURLPrefix):
				a.handleCover(w, r)
			case strings.HasPrefix(r.URL.Path, attachmentURLPrefix):
				a.handleAttachment(w, r)
			case strings.HasPrefix(r.URL.Path, performerImageURLPrefix):
				a.handlePerformerImage(w, r)
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}
