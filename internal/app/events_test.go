package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeScanSubscriber struct {
	ch chan bool
}

func (f *fakeScanSubscriber) ScanSubscribe(ctx context.Context) <-chan bool {
	return f.ch
}

func TestWatchScanEvents_EmitsOnNotify(t *testing.T) {
	sub := &fakeScanSubscriber{ch: make(chan bool, 1)}
	events := make(chan string, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go watchScanEvents(ctx, sub, func(name string, data ...any) {
		events <- name
	})

	sub.ch <- true

	select {
	case name := <-events:
		assert.Equal(t, EventScanComplete, name)
	case <-time.After(time.Second):
		t.Fatal("expected scan:complete event")
	}
}

func TestWatchScanEvents_StopsOnContextCancel(t *testing.T) {
	sub := &fakeScanSubscriber{ch: make(chan bool, 1)}
	done := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		watchScanEvents(ctx, sub, func(string, ...any) {})
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watcher did not stop on context cancel")
	}
}

func TestWatchScanEvents_StopsOnChannelClose(t *testing.T) {
	sub := &fakeScanSubscriber{ch: make(chan bool, 1)}
	done := make(chan struct{})

	go func() {
		watchScanEvents(context.Background(), sub, func(string, ...any) {})
		close(done)
	}()

	close(sub.ch)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watcher did not stop on channel close")
	}
}
