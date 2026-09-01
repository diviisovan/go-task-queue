package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/diviisovan/go-task-queue/internal/handler"
	"github.com/diviisovan/go-task-queue/internal/store"
	"github.com/diviisovan/go-task-queue/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "root@tcp(127.0.0.1:3306)/go_task_queue?parseTime=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Error("database ping failed", "error", err)
		os.Exit(1)
	}

	taskStore := store.New(db)
	if err := taskStore.Migrate(ctx); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}
	logger.Info("database ready", "dsn", "mysql://localhost:3306/go_task_queue")

	w := worker.New(taskStore, logger, 100)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	w.Start(workerCtx)

	pending, err := taskStore.ListPending(context.Background())
	if err != nil {
		logger.Warn("failed to requeue pending tasks", "error", err)
	} else {
		for _, t := range pending {
			w.Enqueue(t)
		}
		if len(pending) > 0 {
			logger.Info("requeued pending tasks from previous run", "count", len(pending))
		}
	}

	h := handler.New(taskStore, w, logger)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	handler.RegisterSwagger(mux)

	logged := loggingMiddleware(logger, mux)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      logged,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		// Caps how long a client may take to send its headers. Without it a
		// client can dribble headers indefinitely and hold the connection open
		// (Slowloris) — ReadTimeout does not cover the header phase alone.
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("shutdown signal received", "signal", sig.String())

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}
	logger.Info("http server stopped")

	workerCancel()
	w.Wait()
	logger.Info("worker stopped, shutdown complete")
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
