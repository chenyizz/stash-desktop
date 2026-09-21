package logger

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

type captureLogger struct {
	LoggerImpl
	messages []string
}

func (l *captureLogger) Debugf(format string, args ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf(format, args...))
}

func TestReadLogMessages(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name: "JSON debug lines",
			input: `{"level":"debug","msg":"one"}` + "\n" +
				`{"level":"debug","msg":"two"}` + "\n",
			want: []string{"one", "two"},
		},
		{
			name:  "no trailing newline",
			input: `{"level":"debug","msg":"last"}`,
			want:  []string{"last"},
		},
		{
			name:  "plain text fallback",
			input: "not a json line\n",
			want:  []string{"not a json line"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &captureLogger{}
			pl := PluginLogger{Logger: l}
			pl.ReadLogMessages(io.NopCloser(strings.NewReader(tt.input)))

			if len(l.messages) != len(tt.want) {
				t.Fatalf("got %d messages, want %d", len(l.messages), len(tt.want))
			}
			for i, want := range tt.want {
				if l.messages[i] != want {
					t.Errorf("message %d: got %.80q, want %.80q",
						i, l.messages[i], want)
				}
			}
		})
	}
}
