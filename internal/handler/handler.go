package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/diviisovan/go-task-queue/internal/model"
)

type TaskStore interface {
	Create(ctx context.Context, req model.CreateTaskRequest) (model.Task, error)
	GetByID(ctx context.Context, id int64) (model.Task, error)
	List(ctx context.Context) ([]model.Task, error)
}

type TaskEnqueuer interface {
	Enqueue(task model.Task) bool
	QueueLen() int
}

type Handler struct {
	store    TaskStore
	enqueuer TaskEnqueuer
	logger   *slog.Logger
}

func New(store TaskStore, enqueuer TaskEnqueuer, logger *slog.Logger) *Handler {
	return &Handler{store: store, enqueuer: enqueuer, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /tasks", h.CreateTask)
	mux.HandleFunc("GET /tasks", h.ListTasks)
	mux.HandleFunc("GET /tasks/{id}", h.GetTask)
	mux.HandleFunc("GET /health", h.Health)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req model.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}

	task, err := h.store.Create(ctx, req)
	if err != nil {
		h.logger.Error("failed to create task", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	if !h.enqueuer.Enqueue(task) {
		h.logger.Warn("task created but queue full", "task_id", task.ID)
	}

	h.logger.Info("task created", "task_id", task.ID, "title", task.Title)
	writeJSON(w, http.StatusCreated, task)
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tasks, err := h.store.List(ctx)
	if err != nil {
		h.logger.Error("failed to list tasks", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if tasks == nil {
		tasks = []model.Task{}
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid task id"})
		return
	}

	task, err := h.store.GetByID(ctx, id)
	if err != nil {
		h.logger.Warn("task not found", "task_id", id, "error", err)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"queue_len": h.enqueuer.QueueLen(),
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
