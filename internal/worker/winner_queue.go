package worker

import (
	"context"
	"sync"
	"time"

	"github.com/producdevity/emuready-discord-giveaway/internal/domain"
	"github.com/rs/zerolog"
)

type WinnerTask struct {
	Interaction domain.Interaction
	Count       int
}

type Runner func(context.Context, WinnerTask) error

type WinnerQueue struct {
	tasks  chan WinnerTask
	runner Runner
	logger zerolog.Logger
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	size   int
	closed bool
	// enqueueWG tracks in-flight enqueue operations to avoid closing the task channel
	// while a sender is mid-blocking send.
	enqueueWG sync.WaitGroup
	mu        sync.Mutex
}

func NewWinnerQueue(ctx context.Context, workers int, runner Runner, logger zerolog.Logger) *WinnerQueue {
	if workers < 1 {
		workers = 1
	}
	c, cancel := context.WithCancel(ctx)
	q := &WinnerQueue{tasks: make(chan WinnerTask, workers*4), runner: runner, logger: logger, ctx: c, cancel: cancel, size: workers}
	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go q.loop()
	}
	return q
}

func (q *WinnerQueue) loop() {
	defer q.wg.Done()
	for {
		select {
		case task, ok := <-q.tasks:
			if !ok {
				return
			}
			rctx, cancel := context.WithTimeout(q.ctx, 2*time.Minute)
			if err := q.runner(rctx, task); err != nil {
				q.logger.Error().Err(err).Msg("winner task failed")
			}
			cancel()
		case <-q.ctx.Done():
			return
		}
	}
}

func (q *WinnerQueue) Enqueue(ctx context.Context, task WinnerTask) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	q.enqueueWG.Add(1)
	defer q.enqueueWG.Done()

	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return context.Canceled
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-q.ctx.Done():
		return q.ctx.Err()
	case q.tasks <- task:
		return nil
	}
}

func (q *WinnerQueue) Close() {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return
	}
	q.closed = true
	q.cancel()
	q.mu.Unlock()
	q.enqueueWG.Wait()
	close(q.tasks)
	q.wg.Wait()
}
