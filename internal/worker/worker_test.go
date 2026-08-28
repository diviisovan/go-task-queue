package worker_test

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/diviisovan/go-task-queue/internal/model"
	"github.com/diviisovan/go-task-queue/internal/worker"
)

type mockStore struct {
	mu      sync.Mutex
	updates []statusUpdate
}

type statusUpdate struct {
	ID     int64
	Status model.TaskStatus
	Result string
}

func (m *mockStore) UpdateStatus(_ context.Context, id int64, status model.TaskStatus, result string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updates = append(m.updates, statusUpdate{ID: id, Status: status, Result: result})
	return nil
}

func (m *mockStore) getUpdates() []statusUpdate {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]statusUpdate, len(m.updates))
	copy(cp, m.updates)
	return cp
}

func TestWorkerProcessesTask(t *testing.T) {
	store := &mockStore{}
	logger := slog.Default()
	w := worker.New(store, logger, 10)

	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)

	task := model.Task{ID: 1, Title: "Test", Description: "hello world"}
	if !w.Enqueue(task) {
		t.Fatal("expected enqueue to succeed")
	}

	time.Sleep(500 * time.Millisecond)
	cancel()
	w.Wait()

	updates := store.getUpdates()
	if len(updates) < 2 {
		t.Fatalf("expected at least 2 updates (processing + completed), got %d", len(updates))
	}
	if updates[0].Status != model.StatusProcessing {
		t.Fatalf("expected first update to be processing, got %s", updates[0].Status)
	}
	if updates[1].Status != model.StatusCompleted {
		t.Fatalf("expected second update to be completed, got %s", updates[1].Status)
	}
}

func TestWorkerDrainsOnShutdown(t *testing.T) {
	store := &mockStore{}
	logger := slog.Default()
	w := worker.New(store, logger, 10)

	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)

	for i := int64(1); i <= 3; i++ {
		w.Enqueue(model.Task{ID: i, Title: "Task", Description: "drain test"})
	}

	time.Sleep(200 * time.Millisecond)
	cancel()
	w.Wait()

	updates := store.getUpdates()
	completed := 0
	for _, u := range updates {
		if u.Status == model.StatusCompleted {
			completed++
		}
	}
	if completed < 1 {
		t.Fatalf("expected at least 1 completed task after drain, got %d", completed)
	}
}

func TestEnqueueRejectsWhenFull(t *testing.T) {
	store := &mockStore{}
	logger := slog.Default()
	w := worker.New(store, logger, 1)

	task := model.Task{ID: 1, Title: "A", Description: "fill"}
	if !w.Enqueue(task) {
		t.Fatal("first enqueue should succeed")
	}
	task2 := model.Task{ID: 2, Title: "B", Description: "overflow"}
	if w.Enqueue(task2) {
		t.Fatal("second enqueue should fail when buffer is full")
	}
}

func TestQueueLen(t *testing.T) {
	store := &mockStore{}
	logger := slog.Default()
	w := worker.New(store, logger, 5)

	if w.QueueLen() != 0 {
		t.Fatalf("expected queue len 0, got %d", w.QueueLen())
	}
	w.Enqueue(model.Task{ID: 1, Title: "A"})
	if w.QueueLen() != 1 {
		t.Fatalf("expected queue len 1, got %d", w.QueueLen())
	}
}
