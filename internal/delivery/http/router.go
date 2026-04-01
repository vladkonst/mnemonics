package http

import (
	"net/http"

	"github.com/rs/zerolog"
	"github.com/vladkonst/mnemonics/internal/delivery/http/handlers"
	"github.com/vladkonst/mnemonics/internal/delivery/http/middleware"
	"github.com/vladkonst/mnemonics/internal/delivery/http/respond"
)

// NewRouter builds and returns the main HTTP handler with all routes and middleware applied.
func NewRouter(
	userH *handlers.UserHandler,
	contentH *handlers.ContentHandler,
	progressH *handlers.ProgressHandler,
	subscriptionH *handlers.SubscriptionHandler,
	paymentH *handlers.PaymentHandler,
	teacherH *handlers.TeacherHandler,
	adminH *handlers.AdminHandler,
	adminToken string,
	log zerolog.Logger,
) http.Handler {
	mux := http.NewServeMux()

	// ── Public routes ────────────────────────────────────────────────────────
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Serve uploaded images (must be registered before the catch-all /api/v1/ handler)
	mux.HandleFunc("GET /api/v1/uploads/{filename}", adminH.ServeUpload)

	// Payment webhook — public (verified internally by signature)
	mux.HandleFunc("POST /api/v1/webhooks/payment-gateway", paymentH.HandleWebhook)

	// ── User-authenticated routes ────────────────────────────────────────────
	authMux := http.NewServeMux()

	// User management
	authMux.HandleFunc("POST /api/v1/users", userH.RegisterUser)
	authMux.HandleFunc("GET /api/v1/users/{user_id}", userH.GetUser)
	authMux.HandleFunc("PATCH /api/v1/users/{user_id}", userH.UpdateUser)
	authMux.HandleFunc("GET /api/v1/users/{user_id}/subscription", userH.GetSubscription)

	// Content
	authMux.HandleFunc("GET /api/v1/content/modules", contentH.GetModules)
	authMux.HandleFunc("GET /api/v1/content/modules/{module_id}", contentH.GetModule)
	authMux.HandleFunc("GET /api/v1/content/modules/{module_id}/themes", contentH.GetModuleThemes)
	authMux.HandleFunc("GET /api/v1/content/themes/{theme_id}", contentH.GetTheme)
	authMux.HandleFunc("POST /api/v1/users/{user_id}/study-sessions", contentH.CreateStudySession)
	authMux.HandleFunc("POST /api/v1/users/{user_id}/test-attempts", contentH.StartTestAttempt)
	authMux.HandleFunc("PUT /api/v1/users/{user_id}/test-attempts/{attempt_id}", contentH.SubmitTestAttempt)
	authMux.HandleFunc("GET /api/v1/users/{user_id}/themes/{theme_id}/access", contentH.CheckThemeAccess)

	// Progress
	authMux.HandleFunc("GET /api/v1/users/{user_id}/progress", progressH.GetUserProgress)
	authMux.HandleFunc("GET /api/v1/users/{user_id}/progress/modules/{module_id}", progressH.GetModuleProgress)

	// Subscription / promo codes
	authMux.HandleFunc("POST /api/v1/teachers/{teacher_id}/promo-codes", subscriptionH.ActivatePromoCode)
	authMux.HandleFunc("GET /api/v1/teachers/{teacher_id}/promo-codes", subscriptionH.GetTeacherPromoCodes)
	authMux.HandleFunc("POST /api/v1/users/{user_id}/subscriptions", subscriptionH.CreateSubscription)

	// Payment invoices
	authMux.HandleFunc("POST /api/v1/users/{user_id}/payment-invoices", paymentH.CreateInvoice)
	authMux.HandleFunc("GET /api/v1/users/{user_id}/payment-invoices/pending", paymentH.GetPendingInvoice)

	// Teacher
	authMux.HandleFunc("GET /api/v1/teachers/{teacher_id}/students", teacherH.GetStudents)
	authMux.HandleFunc("GET /api/v1/teachers/{teacher_id}/students/{student_id}/progress", teacherH.GetStudentProgress)
	authMux.HandleFunc("GET /api/v1/teachers/{teacher_id}/statistics", teacherH.GetStatistics)

	// Apply TelegramAuth middleware to authMux and mount on main mux.
	telegramAuthHandler := middleware.TelegramAuth()(middleware.MaxBody(authMux))
	mux.Handle("/api/v1/", telegramAuthHandler)

	// ── Admin routes ─────────────────────────────────────────────────────────
	adminMux := http.NewServeMux()

	adminMux.HandleFunc("POST /api/v1/admin/upload", adminH.UploadImage)
	adminMux.HandleFunc("POST /api/v1/admin/promo-codes", adminH.CreatePromoCode)
	adminMux.HandleFunc("GET /api/v1/admin/promo-codes", adminH.GetAdminPromoCodes)
	adminMux.HandleFunc("DELETE /api/v1/admin/promo-codes/{code}", adminH.DeactivatePromoCode)
	adminMux.HandleFunc("POST /api/v1/admin/content/modules", adminH.CreateModule)
	adminMux.HandleFunc("GET /api/v1/admin/content/modules", adminH.GetAdminModules)
	adminMux.HandleFunc("GET /api/v1/admin/content/modules/{id}", adminH.GetAdminModule)
	adminMux.HandleFunc("PUT /api/v1/admin/content/modules/{id}", adminH.UpdateModule)
	adminMux.HandleFunc("DELETE /api/v1/admin/content/modules/{id}", adminH.DeleteModule)
	adminMux.HandleFunc("POST /api/v1/admin/content/themes", adminH.CreateTheme)
	adminMux.HandleFunc("GET /api/v1/admin/content/themes", adminH.GetAdminThemes)
	adminMux.HandleFunc("GET /api/v1/admin/content/themes/{id}", adminH.GetAdminTheme)
	adminMux.HandleFunc("PUT /api/v1/admin/content/themes/{id}", adminH.UpdateTheme)
	adminMux.HandleFunc("DELETE /api/v1/admin/content/themes/{id}", adminH.DeleteTheme)
	adminMux.HandleFunc("POST /api/v1/admin/content/mnemonics", adminH.CreateMnemonic)
	adminMux.HandleFunc("GET /api/v1/admin/content/mnemonics", adminH.GetAdminMnemonics)
	adminMux.HandleFunc("GET /api/v1/admin/content/mnemonics/{id}", adminH.GetAdminMnemonic)
	adminMux.HandleFunc("PUT /api/v1/admin/content/mnemonics/{id}", adminH.UpdateMnemonic)
	adminMux.HandleFunc("DELETE /api/v1/admin/content/mnemonics/{id}", adminH.DeleteMnemonic)
	adminMux.HandleFunc("POST /api/v1/admin/content/tests", adminH.CreateTest)
	adminMux.HandleFunc("GET /api/v1/admin/content/tests", adminH.GetAdminTests)
	adminMux.HandleFunc("GET /api/v1/admin/content/tests/{id}", adminH.GetAdminTest)
	adminMux.HandleFunc("PUT /api/v1/admin/content/tests/{id}", adminH.UpdateTest)
	adminMux.HandleFunc("DELETE /api/v1/admin/content/tests/{id}", adminH.DeleteTest)
	adminMux.HandleFunc("POST /api/v1/admin/users", adminH.CreateAdminUser)
	adminMux.HandleFunc("GET /api/v1/admin/users", adminH.GetUsers)
	adminMux.HandleFunc("GET /api/v1/admin/users/{telegram_id}", adminH.GetAdminUser)
	adminMux.HandleFunc("PUT /api/v1/admin/users/{telegram_id}", adminH.UpdateAdminUser)
	adminMux.HandleFunc("GET /api/v1/admin/analytics/overview", adminH.GetAnalytics)

	// Apply AdminAuth middleware then wrap with CORS.
	// CORS is outermost so OPTIONS preflight bypasses auth.
	adminWithAuth := middleware.AdminAuth(adminToken)(adminMux)
	adminWithCORS := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Admin-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		adminWithAuth.ServeHTTP(w, r)
	})
	mux.Handle("/api/v1/admin/", adminWithCORS)

	// ── Global middleware chain ──────────────────────────────────────────────
	var handler http.Handler = mux
	handler = middleware.ContentType()(handler)
	handler = middleware.Logger(log)(handler)
	handler = middleware.RequestID()(handler)
	handler = middleware.Recovery(log)(handler)

	return handler
}
