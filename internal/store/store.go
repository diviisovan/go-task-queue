package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/diviisovan/go-task-queue/internal/model"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Migrate(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS tasks (
		id          BIGINT AUTO_INCREMENT PRIMARY KEY,
		title       VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		status      VARCHAR(20) NOT NULL DEFAULT 'pending',
		result      TEXT NOT NULL,
		created_at  DATETIME(3) NOT NULL,
		updated_at  DATETIME(3) NOT NULL
	)`
	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *Store) Create(ctx context.Context, req model.CreateTaskRequest) (model.Task, error) {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO tasks (title, description, status, result, created_at, updated_at) VALUES (?, ?, ?, '', ?, ?)`,
		req.Title, req.Description, model.StatusPending, now, now,
	)
	if err != nil {
		return model.Task{}, fmt.Errorf("insert task: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.Task{}, fmt.Errorf("last insert id: %w", err)
	}
	return model.Task{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Status:      model.StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *Store) GetByID(ctx context.Context, id int64) (model.Task, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, title, description, status, result, created_at, updated_at FROM tasks WHERE id = ?`, id,
	)
	var t model.Task
	err := row.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Result, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return model.Task{}, fmt.Errorf("get task %d: %w", id, err)
	}
	return t, nil
}

func (s *Store) List(ctx context.Context) ([]model.Task, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, description, status, result, created_at, updated_at FROM tasks ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Result, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (s *Store) UpdateStatus(ctx context.Context, id int64, status model.TaskStatus, result string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE tasks SET status = ?, result = ?, updated_at = ? WHERE id = ?`,
		status, result, time.Now().UTC(), id,
	)
	if err != nil {
		return fmt.Errorf("update task %d: %w", id, err)
	}
	return nil
}

func (s *Store) ListPending(ctx context.Context) ([]model.Task, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, description, status, result, created_at, updated_at FROM tasks WHERE status = ? ORDER BY created_at ASC`,
		model.StatusPending,
	)
	if err != nil {
		return nil, fmt.Errorf("list pending: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Result, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
