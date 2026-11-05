package ingest

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

type EventHandler interface {
	Handle(ctx context.Context, event *ValidatedEvent) error
	HandleDeadLetter(ctx context.Context, raw []byte, reason string) error
}

type Job struct {
	Raw        []byte
	ReceivedAt time.Time
}

type WorkerPool struct {
	workers    int
	jobQueue   chan Job
	handler    EventHandler
	logger     *slog.Logger
	wg         sync.WaitGroup
	running    atomic.Bool
	droppedCnt atomic.Uint64
}

func NewWorkerPool(workers int, queueCapacity int, handler EventHandler, logger *slog.Logger) *WorkerPool {
	if workers <= 0 {
		workers = 16
	}
	if queueCapacity <= 0 {
		queueCapacity = 4096
	}

	return &WorkerPool{
		workers:  workers,
		jobQueue: make(chan Job, queueCapacity),
		handler:  handler,
		logger:   logger,
	}
}

func (p *WorkerPool) Start(ctx context.Context) error {
	p.running.Store(true)
	p.logger.Info("Starting ingestion bounded worker pool", "workers", p.workers, "queue_cap", cap(p.jobQueue))

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		workerID := i
		go func() {
			defer p.wg.Done()
			p.runWorker(ctx, workerID)
		}()
	}

	<-ctx.Done()
	p.running.Store(false)
	close(p.jobQueue)
	p.wg.Wait()
	p.logger.Info("Ingestion worker pool cleanly stopped")
	return nil
}

func (p *WorkerPool) Submit(raw []byte) bool {
	if !p.running.Load() {
		return false
	}

	job := Job{
		Raw:        raw,
		ReceivedAt: time.Now().UTC(),
	}

	select {
	case p.jobQueue <- job:
		return true
	default:
		// Queue full: apply backpressure
		p.droppedCnt.Add(1)
		p.logger.Warn("Ingestion queue backpressure: job queue full", "dropped_total", p.droppedCnt.Load())
		return false
	}
}

func (p *WorkerPool) runWorker(ctx context.Context, workerID int) {
	for job := range p.jobQueue {
		valEv, err := ValidateEvent(job.Raw)
		if err != nil {
			p.logger.Warn("Event validation failed", "worker_id", workerID, "error", err)
			if p.handler != nil {
				_ = p.handler.HandleDeadLetter(ctx, job.Raw, err.Error())
			}
			continue
		}

		if p.handler != nil {
			if err := p.handler.Handle(ctx, valEv); err != nil && !errors.Is(err, context.Canceled) {
				p.logger.Error("Event processing failed", "event_id", valEv.Envelope.EventID, "error", err)
			}
		}
	}
}

func (p *WorkerPool) QueueDepth() int {
	return len(p.jobQueue)
}
