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

	deliveryHTTP "github.com/vladkonst/mnemonics/internal/delivery/http"
	"github.com/vladkonst/mnemonics/internal/delivery/http/handlers"
	"github.com/vladkonst/mnemonics/internal/infrastructure/stub"
	"github.com/vladkonst/mnemonics/internal/repository/sqlite"
	adminUC "github.com/vladkonst/mnemonics/internal/usecase/admin"
	contentUC "github.com/vladkonst/mnemonics/internal/usecase/content"
	paymentUC "github.com/vladkonst/mnemonics/internal/usecase/payment"
	progressUC "github.com/vladkonst/mnemonics/internal/usecase/progress"
	subscriptionUC "github.com/vladkonst/mnemonics/internal/usecase/subscription"
	teacherUC "github.com/vladkonst/mnemonics/internal/usecase/teacher"
	userUC "github.com/vladkonst/mnemonics/internal/usecase/user"
	"github.com/vladkonst/mnemonics/pkg/logger"
)

func main() {
	// Load .env if present (ignore error in production where env vars are set directly).
	_ = godotenv.Load()

	cfg := deliveryHTTP.LoadConfig()

	log := logger.New(cfg.LogLevel, cfg.LogFormat)

	// ── Database ─────────────────────────────────────────────────────────────
	ctx := context.Background()
	db, err := sqlite.Open(ctx, cfg.DBPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}
	defer db.Close()

	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo := sqlite.NewUserRepo(db)
	moduleRepo := sqlite.NewModuleRepo(db)
	themeRepo := sqlite.NewThemeRepo(db)
	mnemonicRepo := sqlite.NewMnemonicRepo(db)
	testRepo := sqlite.NewTestRepo(db)
	progressRepo := sqlite.NewProgressRepo(db)
	attemptRepo := sqlite.NewTestAttemptRepo(db)
	promoCodeRepo := sqlite.NewPromoCodeRepo(db)
	subscriptionRepo := sqlite.NewSubscriptionRepo(db)
	teacherStudentRepo := sqlite.NewTeacherStudentRepo(db)

	// ── External Service Stubs ───────────────────────────────────────────────
	storageSvc := stub.NewStorageService()
	paymentSvc := stub.NewPaymentService()
	notificationSvc := stub.NewNotificationService()

	// ── Use Cases ────────────────────────────────────────────────────────────
	userUseCase := userUC.NewUseCase(userRepo, subscriptionRepo)

	contentUseCase := contentUC.NewUseCase(
		moduleRepo,
		themeRepo,
		mnemonicRepo,
		testRepo,
		progressRepo,
		attemptRepo,
		subscriptionRepo,
		storageSvc,
	)

	progressUseCase := progressUC.NewUseCase(
		progressRepo,
		attemptRepo,
		testRepo,
		themeRepo,
		moduleRepo,
	)

	subscriptionUseCase := subscriptionUC.NewUseCase(
		promoCodeRepo,
		subscriptionRepo,
		userRepo,
		teacherStudentRepo,
		notificationSvc,
	)

	paymentUseCase := paymentUC.NewUseCase(
		userRepo,
		subscriptionRepo,
		paymentSvc,
		notificationSvc,
	)

	teacherUseCase := teacherUC.NewUseCase(
		teacherStudentRepo,
		progressRepo,
		attemptRepo,
		moduleRepo,
		themeRepo,
		userRepo,
	)

	adminUseCase := adminUC.NewUseCase(
		moduleRepo,
		themeRepo,
		mnemonicRepo,
		testRepo,
		promoCodeRepo,
		userRepo,
	)

	// ── Handlers ─────────────────────────────────────────────────────────────
	userHandler := handlers.NewUserHandler(userUseCase)
	contentHandler := handlers.NewContentHandler(contentUseCase)
	progressHandler := handlers.NewProgressHandler(progressUseCase)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionUseCase)
	paymentHandler := handlers.NewPaymentHandler(paymentUseCase)
	teacherHandler := handlers.NewTeacherHandler(teacherUseCase)
	adminHandler := handlers.NewAdminHandler(adminUseCase)

	// ── Router ───────────────────────────────────────────────────────────────
	router := deliveryHTTP.NewRouter(
		userHandler,
		contentHandler,
		progressHandler,
		subscriptionHandler,
		paymentHandler,
		teacherHandler,
		adminHandler,
		cfg.AdminToken,
		log,
	)

	// ── HTTP Server ──────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("addr", cfg.Addr).Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	<-quit
	log.Info().Msg("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server stopped")
}
