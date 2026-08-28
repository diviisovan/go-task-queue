package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"

	"github.com/diviisovan/go-task-queue/internal/handler"
	"github.com/diviisovan/go-task-queue/internal/model"
)

type mockStore struct {
	tasks  []model.Task
	nextID int64
}

func newMockStore() *mockStore {
	return &mockStore{nextID: 1}
}

func (m *mockStore) Create(_ context.Context, req model.CreateTaskRequest) (model.Task, error) {
	t := model.Task{
		ID:          m.nextID,
		Title:       req.Title,
		Description: req.Description,
		Status:      model.StatusPending,
	}
	m.nextID++
	m.tasks = append(m.tasks, t)
	return t, nil
}

func (m *mockStore) GetByID(_ context.Context, id int64) (model.Task, error) {
	for _, t := range m.tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return model.Task{}, context.DeadlineExceeded
}

func (m *mockStore) List(_ context.Context) ([]model.Task, error) {
	return m.tasks, nil
}

type mockEnqueuer struct {
	queued []model.Task
}

func (e *mockEnqueuer) Enqueue(t model.Task) bool {
	e.queued = append(e.queued, t)
	return true
}

func (e *mockEnqueuer) QueueLen() int { return len(e.queued) }

func setup() (*handler.Handler, *mockStore, *mockEnqueuer, *http.ServeMux) {
	store := newMockStore()
	enq := &mockEnqueuer{}
	logger := slog.Default()
	h := handler.New(store, enq, logger)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return h, store, enq, mux
}

func TestCreateTask(t *testing.T) {
	_, _, enq, mux := setup()

	body, _ := json.Marshal(model.CreateTaskRequest{Title: "Test task", Description: "Do something"})
	req := httptest.NewRequest("POST", "/tasks", bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var task model.Task
	json.NewDecoder(w.Body).Decode(&task)
	if task.Title != "Test task" {
		t.Fatalf("expected title 'Test task', got %q", task.Title)
	}
	if task.Status != model.StatusPending {
		t.Fatalf("expected status pending, got %q", task.Status)
	}
	if len(enq.queued) != 1 {
		t.Fatalf("expected 1 enqueued task, got %d", len(enq.queued))
	}
}

func TestCreateTaskMissingTitle(t *testing.T) {
	_, _, _, mux := setup()

	body, _ := json.Marshal(model.CreateTaskRequest{Description: "no title"})
	req := httptest.NewRequest("POST", "/tasks", bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListTasks(t *testing.T) {
	_, store, _, mux := setup()

	store.Create(context.Background(), model.CreateTaskRequest{Title: "A"})
	store.Create(context.Background(), model.CreateTaskRequest{Title: "B"})

	req := httptest.NewRequest("GET", "/tasks", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var tasks []model.Task
	json.NewDecoder(w.Body).Decode(&tasks)
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestGetTask(t *testing.T) {
	_, store, _, mux := setup()

	store.Create(context.Background(), model.CreateTaskRequest{Title: "Find me"})

	req := httptest.NewRequest("GET", "/tasks/1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var task model.Task
	json.NewDecoder(w.Body).Decode(&task)
	if task.Title != "Find me" {
		t.Fatalf("expected 'Find me', got %q", task.Title)
	}
}

func TestGetTaskNotFound(t *testing.T) {
	_, _, _, mux := setup()

	req := httptest.NewRequest("GET", "/tasks/999", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHealth(t *testing.T) {
	_, _, _, mux := setup()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", resp["status"])
	}
}
