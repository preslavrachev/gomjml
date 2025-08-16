package mjml

import (
	"testing"
	"time"
)

// helper to clear singleflight state between tests
func resetSingleflight() {
	sfMutex.Lock()
	sfCalls = make(map[uint64]*sfCall)
	sfMutex.Unlock()
}

func TestSingleflightDoPanicCleanup(t *testing.T) {
	resetSingleflight()

	hash := uint64(42)
	start := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer func() { _ = recover() }()
		singleflightDo(hash, func() (*MJMLNode, error) {
			<-start
			panic("boom")
		})
	}()

	time.Sleep(10 * time.Millisecond) // allow first call to register

	go func() {
		_, _ = singleflightDo(hash, func() (*MJMLNode, error) {
			t.Fatal("second call should not execute")
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
		t.Fatal("singleflightDo did not unblock after panic")
	}
}
