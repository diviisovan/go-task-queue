package store_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/diviisovan/go-task-queue/internal/model"
	"github.com/diviisovan/go-task-queue/internal/store"
)

func testDSN() string {
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		dsn = "root@tcp(127.0.0.1:3306)/go_task_queue_test?parseTime=true"
	}
	return dsn
}

func setupTestDB(t *testing.T) *store.Store {
	t.Helper()
	db, err := sql.Open("mysql", testDSN())
	if err != nil {
		t.Skipf("mysql not available: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("mysql not reachable: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DROP TABLE IF EXISTS tasks")
		db.Close()
	})

	db.Exec("DROP TABLE IF EXISTS tasks")

	s := store.New(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return s
}

func TestCreateAndGet(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	task, err := s.Create(ctx, model.CreateTaskRequest{Title: "Test", Description: "desc"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if task.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if task.Status != model.StatusPending {
		t.Fatalf("expected pending, got %s", task.Status)
	}

	got, err := s.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "Test" {
		t.Fatalf("expected 'Test', got %q", got.Title)
	}
}

func TestList(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	s.Create(ctx, model.CreateTaskRequest{Title: "A"})
	s.Create(ctx, model.CreateTaskRequest{Title: "B"})

	tasks, err := s.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2, got %d", len(tasks))
	}
}

func TestUpdateStatus(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	task, _ := s.Create(ctx, model.CreateTaskRequest{Title: "Update me"})

	err := s.UpdateStatus(ctx, task.ID, model.StatusCompleted, "done")
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	got, _ := s.GetByID(ctx, task.ID)
	if got.Status != model.StatusCompleted {
		t.Fatalf("expected completed, got %s", got.Status)
	}
	if got.Result != "done" {
		t.Fatalf("expected 'done', got %q", got.Result)
	}
}

func TestListPending(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	s.Create(ctx, model.CreateTaskRequest{Title: "Pending"})
	task2, _ := s.Create(ctx, model.CreateTaskRequest{Title: "Done"})
	s.UpdateStatus(ctx, task2.ID, model.StatusCompleted, "ok")

	pending, err := s.ListPending(ctx)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(pending))
	}
	if pending[0].Title != "Pending" {
		t.Fatalf("expected 'Pending', got %q", pending[0].Title)
	}
}

func TestGetNotFound(t *testing.T) {
	s := setupTestDB(t)
	_, err := s.GetByID(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for missing task")
	}
}
