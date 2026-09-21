package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// PluginLogEntry 是插件通过 stderr 发出的 JSON 日志行
type PluginLogEntry struct {
	Level    string  `json:"level"`              // trace/debug/info/warning/error/progress
	Msg      string  `json:"msg,omitempty"`      // 日志消息
	Progress float64 `json:"progress,omitempty"` // 仅 progress 级别使用，0.0 ~ 1.0
}

// PluginLogger 读取插件 stderr 输出，把 JSON 日志转发到主程序 logger
type PluginLogger struct {
	// Logger 是主程序的日志实现
	Logger LoggerImpl
	// Prefix 是插件名前缀，比如 "[myplugin] "
	Prefix string
	// ProgressChan 接收进度值（可选，nil 表示忽略进度）
	ProgressChan chan float64
}

// ReadLogMessages 持续读取插件 stderr 并转发。
// 该方法会阻塞直到 src 关闭或出错，返回前会关闭 src。
func (log *PluginLogger) ReadLogMessages(src io.ReadCloser) {
	defer src.Close()

	reader := bufio.NewReader(src)

	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		if line != "" {
			log.handleLine(line)
		}

		if err != nil {
			// EOF 或其他读取错误，退出
			return
		}
	}
}

// handleLine 处理一行日志。优先按 JSON 解析，失败则当作纯文本处理。
func (log *PluginLogger) handleLine(line string) {
	if log.Logger == nil {
		return
	}

	var entry PluginLogEntry
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		// 不是 JSON，当普通 info 日志
		log.Logger.Infof("%s%s", log.Prefix, line)
		return
	}

	// 空 level 也当 info
	if entry.Level == "" {
		entry.Level = "info"
	}

	switch entry.Level {
	case "trace":
		log.Logger.Tracef("%s%s", log.Prefix, entry.Msg)
	case "debug":
		log.Logger.Debugf("%s%s", log.Prefix, entry.Msg)
	case "info":
		log.Logger.Infof("%s%s", log.Prefix, entry.Msg)
	case "warning", "warn":
		log.Logger.Warnf("%s%s", log.Prefix, entry.Msg)
	case "error":
		log.Logger.Errorf("%s%s", log.Prefix, entry.Msg)
	case "progress":
		if log.ProgressChan != nil {
			// 非阻塞发送，避免插件写得太快阻塞
			select {
			case log.ProgressChan <- entry.Progress:
			default:
			}
		}
	default:
		// 未知 level，当 info
		log.Logger.Infof("%s%s", log.Prefix, entry.Msg)
	}
}

// 保留 Fatal 相关能力（可选）
var _ = fmt.Sprintf // 占位，避免 import 未使用
