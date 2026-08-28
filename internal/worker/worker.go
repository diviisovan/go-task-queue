package worker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/diviisovan/go-task-queue/internal/model"
)

type TaskStore interface {
	UpdateStatus(ctx context.Context, id int64, status model.TaskStatus, result string) error
}

type Worker struct {
	queue  chan model.Task
	store  TaskStore
	logger *slog.Logger
	wg     sync.WaitGroup
}

func New(store TaskStore, logger *slog.Logger, bufSize int) *Worker {
	return &Worker{
		queue:  make(chan model.Task, bufSize),
		store:  store,
		logger: logger,
	}
}

func (w *Worker) Enqueue(task model.Task) bool {
	select {
	case w.queue <- task:
		w.logger.Info("task enqueued", "task_id", task.ID, "title", task.Title)
		return true
	default:
		w.logger.Warn("queue full, rejecting task", "task_id", task.ID)
		return false
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.logger.Info("worker started")
		for {
			select {
			case <-ctx.Done():
				w.logger.Info("worker shutting down, draining queue")
				w.drain()
				return
			case task := <-w.queue:
				w.process(ctx, task)
			}
		}
	}()
}

func (w *Worker) Wait() {
	w.wg.Wait()
}

func (w *Worker) QueueLen() int {
	return len(w.queue)
}

func (w *Worker) process(ctx context.Context, task model.Task) {
	w.logger.Info("processing task", "task_id", task.ID, "title", task.Title)

	jobCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := w.store.UpdateStatus(jobCtx, task.ID, model.StatusProcessing, ""); err != nil {
		w.logger.Error("failed to mark processing", "task_id", task.ID, "error", err)
		return
	}

	result, err := w.execute(task)
	if err != nil {
		w.logger.Error("task failed", "task_id", task.ID, "error", err)
		_ = w.store.UpdateStatus(jobCtx, task.ID, model.StatusFailed, err.Error())
		return
	}

	if err := w.store.UpdateStatus(jobCtx, task.ID, model.StatusCompleted, result); err != nil {
		w.logger.Error("failed to mark completed", "task_id", task.ID, "error", err)
		return
	}
	w.logger.Info("task completed", "task_id", task.ID, "result", result)
}

func (w *Worker) execute(task model.Task) (string, error) {
	time.Sleep(100 * time.Millisecond)
	wordCount := len(strings.Fields(task.Description))
	return fmt.Sprintf("processed: %d words in description", wordCount), nil
}

func (w *Worker) drain() {
	for {
		select {
		case task := <-w.queue:
			w.logger.Info("draining task", "task_id", task.ID)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			w.process(ctx, task)
			cancel()
		default:
			w.logger.Info("queue drained")
			return
		}
	}
}
