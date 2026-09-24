package mjml

import (
	"testing"
	"time"
)

func TestSingleflightDoPanicCleanup(t *testing.T) {
	var g singleflightGroup

	hash := uint64(42)
	start := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer func() { _ = recover() }()
		g.do(hash, func() (*MJMLNode, error) {
			<-start
			panic("boom")
		})
	}()

	time.Sleep(10 * time.Millisecond) // allow first call to register

	go func() {
		_, _ = g.do(hash, func() (*MJMLNode, error) {
			t.Error("second call should not execute")
			return nil, nil
		})
		close(done)
	}()

	time.Sleep(10 * time.Millisecond) // allow second call to block
	close(start)

	select {
	case <-done:
		// success: second call returned
	case <-time.After(time.Second):
		t.Fatal("singleflight did not unblock after panic")
	}
}
