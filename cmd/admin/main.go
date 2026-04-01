package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	admin_panel "github.com/vladkonst/mnemonics/internal/admin_panel"
	"github.com/vladkonst/mnemonics/internal/repository/sqlite"
)

func main() {
	_ = godotenv.Load()

	dbPath := os.Getenv("SQLITE_PATH")
	if dbPath == "" {
		dbPath = "mnemo.db"
	}
	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		adminToken = "changeme"
	}
	addr := os.Getenv("ADMIN_ADDR")
	if addr == "" {
		addr = ":9000"
	}

	ctx := context.Background()
	db, err := sqlite.Open(ctx, dbPath)
	if err != nil {
		panic("failed to open database: " + err.Error())
	}
	defer db.Close()

	srv, err := admin_panel.NewServer(db, adminToken)
	if err != nil {
		panic("failed to create admin server: " + err.Error())
	}

	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      srv.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic("server failed: " + err.Error())
		}
	}()

	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}
