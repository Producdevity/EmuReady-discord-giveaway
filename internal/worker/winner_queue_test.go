package worker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestWinnerQueueCancelsInFlightTasksOnClose(t *testing.T) {
	queueCtx, queueCancel := context.WithCancel(context.Background())
	defer queueCancel()

	runnerDone := make(chan struct{})
	taskStarted := make(chan struct{})

	queue := NewWinnerQueue(queueCtx, 1, func(ctx context.Context, task WinnerTask) error {
		close(taskStarted)
		<-ctx.Done()
		close(runnerDone)
		return ctx.Err()
	}, zerolog.Nop())

	if err := queue.Enqueue(context.Background(), WinnerTask{Count: 1}); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	select {
	case <-taskStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("runner did not start")
	}

	queue.Close()

	select {
	case <-runnerDone:
	case <-time.After(2 * time.Second):
		t.Fatal("runner did not observe cancellation")
	}
}

func TestWinnerQueueReturnsErrorAfterClose(t *testing.T) {
	queue := NewWinnerQueue(context.Background(), 1, func(context.Context, WinnerTask) error {
		return nil
	}, zerolog.Nop())

	if err := queue.Enqueue(context.Background(), WinnerTask{Count: 1}); err != nil {
		t.Fatalf("enqueue failed before close: %v", err)
	}

	queue.Close()

	if err := queue.Enqueue(context.Background(), WinnerTask{Count: 1}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled after close, got: %v", err)
	}
}

func TestWinnerQueueRespectsCallerContextCancellation(t *testing.T) {
	queue := NewWinnerQueue(context.Background(), 1, func(context.Context, WinnerTask) error {
		return nil
	}, zerolog.Nop())
	defer queue.Close()

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := queue.Enqueue(cancelledCtx, WinnerTask{Count: 1}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected caller context cancellation, got: %v", err)
	}
}

func TestWinnerQueueConcurrentCloseAndEnqueueDoesNotPanic(t *testing.T) {
	queue := NewWinnerQueue(context.Background(), 1, func(context.Context, WinnerTask) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}, zerolog.Nop())

	panicCh := make(chan interface{}, 1)
	const workerCount = 200
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					select {
					case panicCh <- r:
					default:
					}
				}
			}()
			_ = queue.Enqueue(context.Background(), WinnerTask{Count: 1})
		}()
	}

	queue.Close()
	wg.Wait()

	select {
	case recovered := <-panicCh:
		t.Fatalf("enqueue panicked: %v", recovered)
	default:
	}
}
