// Package http_test contains end-to-end tests for the full HTTP stack.
// Uses a real in-memory SQLite database and httptest server.
package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/rs/zerolog"
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
)

const testAdminToken = "test-admin-token"

// newTestServer builds a full HTTP server backed by in-memory SQLite.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	ctx := context.Background()
	db, err := sqlite.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

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

	storageSvc := stub.NewStorageService(t.TempDir())
	paymentSvc := stub.NewPaymentService()
	notificationSvc := stub.NewNotificationService()

	userUseCase := userUC.NewUseCase(userRepo, subscriptionRepo)
	contentUseCase := contentUC.NewUseCase(
		moduleRepo, themeRepo, mnemonicRepo, testRepo,
		progressRepo, attemptRepo, subscriptionRepo, storageSvc,
	)
	progressUseCase := progressUC.NewUseCase(
		progressRepo, attemptRepo, testRepo, themeRepo, moduleRepo,
	)
	subscriptionUseCase := subscriptionUC.NewUseCase(
		promoCodeRepo, subscriptionRepo, userRepo, teacherStudentRepo, notificationSvc,
	)
	paymentUseCase := paymentUC.NewUseCase(
		userRepo, subscriptionRepo, paymentSvc, notificationSvc,
	)
	teacherUseCase := teacherUC.NewUseCase(
		teacherStudentRepo, progressRepo, attemptRepo, moduleRepo, themeRepo, userRepo,
	)
	adminUseCase := adminUC.NewUseCase(
		moduleRepo, themeRepo, mnemonicRepo, testRepo, promoCodeRepo, userRepo, db,
	)

	router := deliveryHTTP.NewRouter(
		handlers.NewUserHandler(userUseCase),
		handlers.NewContentHandler(contentUseCase),
		handlers.NewProgressHandler(progressUseCase),
		handlers.NewSubscriptionHandler(subscriptionUseCase),
		handlers.NewPaymentHandler(paymentUseCase),
		handlers.NewTeacherHandler(teacherUseCase),
		handlers.NewAdminHandler(adminUseCase, storageSvc, t.TempDir()),
		testAdminToken,
		zerolog.Nop(),
	)

	return httptest.NewServer(router)
}

func doJSON(t *testing.T, server *httptest.Server, method, path string, body any, headers map[string]string) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req, err := http.NewRequest(method, server.URL+path, &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, dst any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

// ── Health check ─────────────────────────────────────────────────────────────

func TestE2E_Health(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

// ── Auth middleware ───────────────────────────────────────────────────────────

func TestE2E_RequiresTelegramAuth(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/content/modules?user_id=1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestE2E_AdminRequiresToken(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := doJSON(t, srv, http.MethodGet, "/api/v1/admin/users", nil, nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

// ── User registration flow ────────────────────────────────────────────────────

func TestE2E_RegisterUser(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	headers := map[string]string{"X-Telegram-User-Id": "12345"}
	body := map[string]any{
		"telegram_id": 12345,
		
		
		"username":    "ivan_p",
	}

	resp := doJSON(t, srv, http.MethodPost, "/api/v1/users", body, headers)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}

	var result map[string]any
	decodeJSON(t, resp, &result)
	if result["telegram_id"] != float64(12345) {
		t.Errorf("telegram_id = %v, want 12345", result["telegram_id"])
	}
}

func TestE2E_RegisterUser_Duplicate(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	headers := map[string]string{"X-Telegram-User-Id": "999"}
	body := map[string]any{"telegram_id": 999, }

	doJSON(t, srv, http.MethodPost, "/api/v1/users", body, headers)
	resp := doJSON(t, srv, http.MethodPost, "/api/v1/users", body, headers)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("status = %d, want 409", resp.StatusCode)
	}
}

func TestE2E_RegisterUser_MissingFields(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	headers := map[string]string{"X-Telegram-User-Id": "111"}

	// missing telegram_id
	resp := doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{}, headers)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("missing telegram_id: status = %d, want 400", resp.StatusCode)
	}
}

// ── Update user role ──────────────────────────────────────────────────────────

func TestE2E_UpdateUserRole(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	headers := map[string]string{"X-Telegram-User-Id": "500"}

	// Register first
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 500, }, headers)

	// Update role
	resp := doJSON(t, srv, http.MethodPatch, "/api/v1/users/500",
		map[string]any{"role": "teacher"}, headers)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var result map[string]any
	decodeJSON(t, resp, &result)
	if result["role"] != "teacher" {
		t.Errorf("role = %v, want teacher", result["role"])
	}
}

// ── Subscription (no subscription) ───────────────────────────────────────────

func TestE2E_GetSubscription_NotFound(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	headers := map[string]string{"X-Telegram-User-Id": "600"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 600, }, headers)

	resp := doJSON(t, srv, http.MethodGet, "/api/v1/users/600/subscription", nil, headers)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

// ── Content modules (empty) ───────────────────────────────────────────────────

func TestE2E_GetModules_Empty(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	headers := map[string]string{"X-Telegram-User-Id": "700"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 700, }, headers)

	resp := doJSON(t, srv, http.MethodGet, "/api/v1/content/modules?user_id=700", nil, headers)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	var result map[string]any
	decodeJSON(t, resp, &result)
	modules, _ := result["modules"].([]any)
	if len(modules) != 0 {
		t.Errorf("expected empty modules, got %d", len(modules))
	}
}

// ── Admin: create module ──────────────────────────────────────────────────────

func TestE2E_AdminCreateModule(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	headers := map[string]string{"X-Admin-Token": testAdminToken}
	body := map[string]any{
		"name":        "Anatomy",
		"description": "Study of body",
		"order_num":   1,
		"is_locked":   false,
	}

	resp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/modules", body, headers)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}

	var result map[string]any
	decodeJSON(t, resp, &result)
	if result["name"] != "Anatomy" {
		t.Errorf("name = %v, want Anatomy", result["name"])
	}
}

// ── Full study flow ───────────────────────────────────────────────────────────

func TestE2E_StudyFlow_ModuleAndThemes(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminHeaders := map[string]string{"X-Admin-Token": testAdminToken}
	userHeaders := map[string]string{"X-Telegram-User-Id": "800"}

	// Register user
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 800, }, userHeaders)

	// Create module
	modResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/modules",
		map[string]any{"name": "Histology", "description": "Cells", "order_num": 1, "is_locked": false},
		adminHeaders)
	var mod map[string]any
	decodeJSON(t, modResp, &mod)
	idVal, ok := mod["id"]
	if !ok {
		t.Fatalf("module response missing ID field, got: %v", mod)
	}
	modID := int(idVal.(float64))

	// Create theme
	themeResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/themes",
		map[string]any{
			"module_id":    modID,
			"name":         "Cell Structure",
			"description":  "Basic cell anatomy",
			"order_num":    1,
			"is_introduction": true,
			"is_locked":    false,
		},
		adminHeaders)
	if themeResp.StatusCode != http.StatusCreated {
		t.Fatalf("create theme: status = %d, want 201", themeResp.StatusCode)
	}
	themeResp.Body.Close()

	// List modules — should now have one
	resp := doJSON(t, srv, http.MethodGet, "/api/v1/content/modules?user_id=800", nil, userHeaders)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("list modules: status = %d, want 200", resp.StatusCode)
	}
	var modulesResp map[string]any
	decodeJSON(t, resp, &modulesResp)
	modules, _ := modulesResp["modules"].([]any)
	if len(modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(modules))
	}
}

// ── Progress (empty) ──────────────────────────────────────────────────────────

func TestE2E_UserProgress_Empty(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	headers := map[string]string{"X-Telegram-User-Id": "900"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 900, }, headers)

	resp := doJSON(t, srv, http.MethodGet, "/api/v1/users/900/progress", nil, headers)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	var result map[string]any
	decodeJSON(t, resp, &result)
	if result["total_themes"] != float64(0) {
		t.Errorf("total_themes = %v, want 0", result["total_themes"])
	}
}

// ── Admin: promo codes ────────────────────────────────────────────────────────

func TestE2E_AdminCreatePromoCode(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	headers := map[string]string{"X-Admin-Token": testAdminToken}
	body := map[string]any{
		"code":            "UNI2024",
		"university_name": "MSU",
		"max_activations": 10,
	}

	resp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/promo-codes", body, headers)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
	var result map[string]any
	decodeJSON(t, resp, &result)
	if result["code"] != "UNI2024" {
		t.Errorf("code = %v, want UNI2024", result["code"])
	}
}

func TestE2E_AdminDeactivatePromoCode(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}

	doJSON(t, srv, http.MethodPost, "/api/v1/admin/promo-codes",
		map[string]any{"code": "DEL2024", "university_name": "SPbU", "max_activations": 5}, adminH)

	resp := doJSON(t, srv, http.MethodDelete, "/api/v1/admin/promo-codes/DEL2024", nil, adminH)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestE2E_AdminGetUsers(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	// Register a user first
	userH := map[string]string{"X-Telegram-User-Id": "1001"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 1001, }, userH)

	adminH := map[string]string{"X-Admin-Token": testAdminToken}
	resp := doJSON(t, srv, http.MethodGet, "/api/v1/admin/users?limit=10&offset=0", nil, adminH)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_AdminGetAnalytics(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	headers := map[string]string{"X-Admin-Token": testAdminToken}
	resp := doJSON(t, srv, http.MethodGet, "/api/v1/admin/analytics/overview", nil, headers)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Content: themes and study session ────────────────────────────────────────

func TestE2E_GetModuleThemes(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}
	userH := map[string]string{"X-Telegram-User-Id": "1100"}

	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 1100, }, userH)

	modResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/modules",
		map[string]any{"name": "Bio", "description": "Desc", "order_num": 1, "is_locked": false}, adminH)
	var mod map[string]any
	decodeJSON(t, modResp, &mod)
	modID := int(mod["id"].(float64))

	resp := doJSON(t, srv, http.MethodGet,
		"/api/v1/content/modules/"+itoa(modID)+"/themes?user_id=1100", nil, userH)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetModuleThemes: status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_CreateStudySession(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}
	userH := map[string]string{"X-Telegram-User-Id": "1200"}

	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 1200, }, userH)

	modResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/modules",
		map[string]any{"name": "Chem", "description": "D", "order_num": 1, "is_locked": false}, adminH)
	var mod map[string]any
	decodeJSON(t, modResp, &mod)
	modID := int(mod["id"].(float64))

	themeResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/themes",
		map[string]any{
			"module_id": modID, "name": "Intro", "description": "D",
			"order_num": 1, "is_introduction": true, "is_locked": false,
		}, adminH)
	var theme map[string]any
	decodeJSON(t, themeResp, &theme)
	themeID := int(theme["id"].(float64))

	resp := doJSON(t, srv, http.MethodPost, "/api/v1/users/1200/study-sessions",
		map[string]any{"theme_id": themeID}, userH)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("CreateStudySession: status = %d, want 201", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_StartAndSubmitTestAttempt(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}
	userH := map[string]string{"X-Telegram-User-Id": "1300"}

	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 1300, }, userH)

	modResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/modules",
		map[string]any{"name": "Phys", "description": "D", "order_num": 1, "is_locked": false}, adminH)
	var mod map[string]any
	decodeJSON(t, modResp, &mod)
	modID := int(mod["id"].(float64))

	themeResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/themes",
		map[string]any{
			"module_id": modID, "name": "Heart", "description": "D",
			"order_num": 1, "is_introduction": true, "is_locked": false,
		}, adminH)
	var theme map[string]any
	decodeJSON(t, themeResp, &theme)
	themeID := int(theme["id"].(float64))

	// Create test for theme
	doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/tests",
		map[string]any{
			"theme_id": themeID, "difficulty": 2, "passing_score": 70,
			"shuffle_questions": false, "shuffle_answers": false,
			"questions": []map[string]any{
				{
					"id": 1, "text": "What is 2+2?", "type": "multiple_choice",
					"options": []string{"3", "4", "5"}, "correct_answer": "4", "order_num": 1,
				},
			},
		}, adminH)

	// Create study session first (to start theme)
	doJSON(t, srv, http.MethodPost, "/api/v1/users/1300/study-sessions",
		map[string]any{"theme_id": themeID}, userH)

	// Start test attempt
	attemptResp := doJSON(t, srv, http.MethodPost, "/api/v1/users/1300/test-attempts",
		map[string]any{"theme_id": themeID}, userH)
	if attemptResp.StatusCode != http.StatusCreated {
		t.Fatalf("StartTestAttempt: status = %d, want 201", attemptResp.StatusCode)
	}
	var attempt map[string]any
	decodeJSON(t, attemptResp, &attempt)
	attemptID, _ := attempt["attempt_id"].(string)
	if attemptID == "" {
		t.Fatal("AttemptID is empty")
	}

	// Submit answers
	submitResp := doJSON(t, srv, http.MethodPut,
		"/api/v1/users/1300/test-attempts/"+attemptID,
		map[string]any{"answers": []map[string]any{{"question_id": 1, "answer": "4"}}},
		userH)
	if submitResp.StatusCode != http.StatusOK {
		t.Errorf("SubmitTestAttempt: status = %d, want 200", submitResp.StatusCode)
	}
	submitResp.Body.Close()
}

func TestE2E_CheckThemeAccess(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}
	userH := map[string]string{"X-Telegram-User-Id": "1400"}

	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 1400, }, userH)

	modResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/modules",
		map[string]any{"name": "Bio2", "description": "D", "order_num": 1, "is_locked": false}, adminH)
	var mod map[string]any
	decodeJSON(t, modResp, &mod)
	modID := int(mod["id"].(float64))

	themeResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/themes",
		map[string]any{
			"module_id": modID, "name": "T1", "description": "D",
			"order_num": 1, "is_introduction": true, "is_locked": false,
		}, adminH)
	var theme map[string]any
	decodeJSON(t, themeResp, &theme)
	themeID := int(theme["id"].(float64))

	resp := doJSON(t, srv, http.MethodGet,
		"/api/v1/users/1400/themes/"+itoa(themeID)+"/access", nil, userH)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("CheckThemeAccess: status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Progress: module progress ─────────────────────────────────────────────────

func TestE2E_GetModuleProgress(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}
	userH := map[string]string{"X-Telegram-User-Id": "1500"}

	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 1500, }, userH)

	modResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/modules",
		map[string]any{"name": "ModProg", "description": "D", "order_num": 1, "is_locked": false}, adminH)
	var mod map[string]any
	decodeJSON(t, modResp, &mod)
	modID := int(mod["id"].(float64))

	resp := doJSON(t, srv, http.MethodGet,
		"/api/v1/users/1500/progress/modules/"+itoa(modID), nil, userH)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetModuleProgress: status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Subscription: promo code activation ──────────────────────────────────────

func TestE2E_SubscriptionPromoFlow(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}
	teacherH := map[string]string{"X-Telegram-User-Id": "2000"}
	studentH := map[string]string{"X-Telegram-User-Id": "2001"}

	// Register teacher and student
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 2000, }, teacherH)
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 2001, }, studentH)

	// Set teacher role
	doJSON(t, srv, http.MethodPatch, "/api/v1/users/2000",
		map[string]any{"role": "teacher"}, teacherH)

	// Admin creates promo code
	doJSON(t, srv, http.MethodPost, "/api/v1/admin/promo-codes",
		map[string]any{"code": "FLOW2024", "university_name": "TestU", "max_activations": 5}, adminH)

	// Teacher activates promo code
	activateResp := doJSON(t, srv, http.MethodPost, "/api/v1/teachers/2000/promo-codes",
		map[string]any{"code": "FLOW2024"}, teacherH)
	if activateResp.StatusCode != http.StatusOK {
		t.Errorf("ActivatePromoCode: status = %d, want 200", activateResp.StatusCode)
	}
	activateResp.Body.Close()

	// Teacher gets their promo codes
	listResp := doJSON(t, srv, http.MethodGet, "/api/v1/teachers/2000/promo-codes", nil, teacherH)
	if listResp.StatusCode != http.StatusOK {
		t.Errorf("GetTeacherPromoCodes: status = %d, want 200", listResp.StatusCode)
	}
	listResp.Body.Close()

	// Student creates subscription with promo code
	subResp := doJSON(t, srv, http.MethodPost, "/api/v1/users/2001/subscriptions",
		map[string]any{"type": "promo", "promo_code": "FLOW2024"}, studentH)
	if subResp.StatusCode != http.StatusCreated {
		t.Errorf("CreateSubscription: status = %d, want 201", subResp.StatusCode)
	}
	subResp.Body.Close()
}

// ── Payment invoice ───────────────────────────────────────────────────────────

func TestE2E_CreatePaymentInvoice(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "3000"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 3000, }, userH)

	resp := doJSON(t, srv, http.MethodPost, "/api/v1/users/3000/payment-invoices",
		map[string]any{"plan": "monthly"}, userH)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("CreateInvoice: status = %d, want 201", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_GetPendingInvoice(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "3001"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 3001, }, userH)

	resp := doJSON(t, srv, http.MethodGet, "/api/v1/users/3001/payment-invoices/pending", nil, userH)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetPendingInvoice: status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Teacher endpoints ─────────────────────────────────────────────────────────

func TestE2E_TeacherStudents(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}
	teacherH := map[string]string{"X-Telegram-User-Id": "4000"}
	studentH := map[string]string{"X-Telegram-User-Id": "4001"}

	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 4000, }, teacherH)
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 4001, }, studentH)
	doJSON(t, srv, http.MethodPatch, "/api/v1/users/4000",
		map[string]any{"role": "teacher"}, teacherH)

	// Create and activate promo, then student subscribes
	doJSON(t, srv, http.MethodPost, "/api/v1/admin/promo-codes",
		map[string]any{"code": "TEACH001", "university_name": "U", "max_activations": 5}, adminH)
	doJSON(t, srv, http.MethodPost, "/api/v1/teachers/4000/promo-codes",
		map[string]any{"code": "TEACH001"}, teacherH)
	doJSON(t, srv, http.MethodPost, "/api/v1/users/4001/subscriptions",
		map[string]any{"type": "promo", "promo_code": "TEACH001"}, studentH)

	// GetStudents
	studResp := doJSON(t, srv, http.MethodGet, "/api/v1/teachers/4000/students", nil, teacherH)
	if studResp.StatusCode != http.StatusOK {
		t.Errorf("GetStudents: status = %d, want 200", studResp.StatusCode)
	}
	studResp.Body.Close()

	// GetStudentProgress
	progResp := doJSON(t, srv, http.MethodGet,
		"/api/v1/teachers/4000/students/4001/progress", nil, teacherH)
	if progResp.StatusCode != http.StatusOK {
		t.Errorf("GetStudentProgress: status = %d, want 200", progResp.StatusCode)
	}
	progResp.Body.Close()

	// GetStatistics
	statsResp := doJSON(t, srv, http.MethodGet, "/api/v1/teachers/4000/statistics", nil, teacherH)
	if statsResp.StatusCode != http.StatusOK {
		t.Errorf("GetStatistics: status = %d, want 200", statsResp.StatusCode)
	}
	statsResp.Body.Close()
}

// ── Admin: content creation ───────────────────────────────────────────────────

func TestE2E_AdminCreateMnemonicAndUpdateModule(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}

	// Create module
	modResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/modules",
		map[string]any{"name": "Neuro", "description": "D", "order_num": 1, "is_locked": false}, adminH)
	var mod map[string]any
	decodeJSON(t, modResp, &mod)
	modID := int(mod["id"].(float64))

	// Update module
	updateResp := doJSON(t, srv, http.MethodPut, "/api/v1/admin/content/modules/"+itoa(modID),
		map[string]any{"name": "Neurology", "description": "Updated", "order_num": 1, "is_locked": false},
		adminH)
	if updateResp.StatusCode != http.StatusOK {
		t.Errorf("UpdateModule: status = %d, want 200", updateResp.StatusCode)
	}
	updateResp.Body.Close()

	// Create theme
	themeResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/themes",
		map[string]any{
			"module_id": modID, "name": "Neurons", "description": "D",
			"order_num": 1, "is_introduction": true, "is_locked": false,
		}, adminH)
	var theme map[string]any
	decodeJSON(t, themeResp, &theme)
	themeID := int(theme["id"].(float64))

	// Create text mnemonic
	text := "Neurons transmit signals"
	mnemoResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/mnemonics",
		map[string]any{
			"theme_id": themeID, "type": "text",
			"content_text": text, "order_num": 1,
		}, adminH)
	if mnemoResp.StatusCode != http.StatusCreated {
		t.Errorf("CreateMnemonic: status = %d, want 201", mnemoResp.StatusCode)
	}
	mnemoResp.Body.Close()
}

// ── Payment webhook ───────────────────────────────────────────────────────────

func TestE2E_PaymentWebhook_Successful(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "5000"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 5000, }, userH)

	// Create an invoice first so user has a pending payment ID
	invoiceResp := doJSON(t, srv, http.MethodPost, "/api/v1/users/5000/payment-invoices",
		map[string]any{"plan": "monthly"}, userH)
	var invoice map[string]any
	decodeJSON(t, invoiceResp, &invoice)
	invoiceID, _ := invoice["invoice_id"].(string)

	// Send successful payment webhook (public endpoint, no auth)
	webhookBody := map[string]any{
		"payment_id": invoiceID,
		"user_id":    5000,
		"plan":       "monthly",
		"status":     "succeeded",
	}
	resp := doJSON(t, srv, http.MethodPost, "/api/v1/webhooks/payment-gateway",
		webhookBody, map[string]string{})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("HandleWebhook: status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_PaymentWebhook_MalformedBody(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	// Malformed JSON — should still return 200
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/webhooks/payment-gateway",
		bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 even for malformed body", resp.StatusCode)
	}
}

// ── Image mnemonic + study session with presign ───────────────────────────────

func TestE2E_StudySession_WithImageMnemonic(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}
	userH := map[string]string{"X-Telegram-User-Id": "6000"}

	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 6000, }, userH)

	modResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/modules",
		map[string]any{"name": "ImgMod", "description": "D", "order_num": 1, "is_locked": false}, adminH)
	var mod map[string]any
	decodeJSON(t, modResp, &mod)
	modID := int(mod["id"].(float64))

	themeResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/themes",
		map[string]any{
			"module_id": modID, "name": "ImgTheme", "description": "D",
			"order_num": 1, "is_introduction": true, "is_locked": false,
		}, adminH)
	var theme map[string]any
	decodeJSON(t, themeResp, &theme)
	themeID := int(theme["id"].(float64))

	// Create image mnemonic (covers PresignURL in study session)
	imgResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/mnemonics",
		map[string]any{
			"theme_id": themeID, "type": "image",
			"s3_image_key": "images/cell.png", "order_num": 1,
		}, adminH)
	if imgResp.StatusCode != http.StatusCreated {
		t.Fatalf("CreateImageMnemonic: status = %d, want 201", imgResp.StatusCode)
	}
	imgResp.Body.Close()

	// Create study session — triggers PresignURL for image mnemonic
	sessionResp := doJSON(t, srv, http.MethodPost, "/api/v1/users/6000/study-sessions",
		map[string]any{"theme_id": themeID}, userH)
	if sessionResp.StatusCode != http.StatusCreated {
		t.Errorf("CreateStudySession with image: status = %d, want 201", sessionResp.StatusCode)
	}
	sessionResp.Body.Close()
}

// ── CreateSubscription: payment type ─────────────────────────────────────────

func TestE2E_CreateSubscription_Payment(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "7000"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 7000, }, userH)

	// Activate subscription via payment type
	resp := doJSON(t, srv, http.MethodPost, "/api/v1/users/7000/subscriptions",
		map[string]any{"type": "payment", "payment_id": "pay_test_001", "plan": "monthly"}, userH)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("CreateSubscription payment: status = %d, want 201", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_CreateSubscription_Payment_Idempotent(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "7001"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 7001, }, userH)

	body := map[string]any{"type": "payment", "payment_id": "pay_dup_001", "plan": "yearly"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users/7001/subscriptions", body, userH)
	// Second call should be idempotent
	resp := doJSON(t, srv, http.MethodPost, "/api/v1/users/7001/subscriptions", body, userH)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("idempotent payment: status = %d, want 201", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_CreateSubscription_InvalidType(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "7002"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 7002, }, userH)

	resp := doJSON(t, srv, http.MethodPost, "/api/v1/users/7002/subscriptions",
		map[string]any{"type": "invalid"}, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("invalid type: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_CreateSubscription_MissingPaymentID(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "7003"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 7003, }, userH)

	resp := doJSON(t, srv, http.MethodPost, "/api/v1/users/7003/subscriptions",
		map[string]any{"type": "payment"}, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("missing payment_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_CreateSubscription_MissingPromoCode(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "7004"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 7004, }, userH)

	resp := doJSON(t, srv, http.MethodPost, "/api/v1/users/7004/subscriptions",
		map[string]any{"type": "promo"}, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("missing promo_code: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── UpdateUser: settings ──────────────────────────────────────────────────────

func TestE2E_UpdateUser_Settings(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "7010"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 7010, }, userH)

	lang := "ru"
	enabled := true
	resp := doJSON(t, srv, http.MethodPatch, "/api/v1/users/7010",
		map[string]any{"language": lang, "notifications_enabled": enabled}, userH)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("UpdateUser settings: status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_UpdateUser_NoFields(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "7011"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 7011, }, userH)

	resp := doJSON(t, srv, http.MethodPatch, "/api/v1/users/7011",
		map[string]any{}, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("UpdateUser no fields: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestE2E_UpdateUser_BadID(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "7012"}
	resp := doJSON(t, srv, http.MethodPatch, "/api/v1/users/notanumber",
		map[string]any{"role": "teacher"}, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("UpdateUser bad id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Bad path params ───────────────────────────────────────────────────────────

func TestE2E_BadPathParams(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "7020"}
	adminH := map[string]string{"X-Admin-Token": testAdminToken}

	// Bad user_id in GetSubscription
	resp := doJSON(t, srv, http.MethodGet, "/api/v1/users/bad/subscription", nil, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("GetSubscription bad id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// Bad teacher_id
	resp = doJSON(t, srv, http.MethodGet, "/api/v1/teachers/bad/promo-codes", nil, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("GetTeacherPromoCodes bad id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// Bad teacher_id in ActivatePromoCode
	resp = doJSON(t, srv, http.MethodPost, "/api/v1/teachers/bad/promo-codes",
		map[string]any{"code": "X"}, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("ActivatePromoCode bad teacher_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// Bad user_id in CreateSubscription
	resp = doJSON(t, srv, http.MethodPost, "/api/v1/users/bad/subscriptions",
		map[string]any{"type": "promo", "promo_code": "X"}, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateSubscription bad user_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// Bad module_id in GetModuleThemes
	resp = doJSON(t, srv, http.MethodGet, "/api/v1/content/modules/bad/themes?user_id=1", nil, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("GetModuleThemes bad module_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// GetUsers with admin (valid)
	resp = doJSON(t, srv, http.MethodGet, "/api/v1/admin/users?limit=bad", nil, adminH)
	// Should succeed with default limit
	resp.Body.Close()

	// Admin DeactivatePromoCode for nonexistent code — SQLite UPDATE with no rows returns no error → 200
	resp = doJSON(t, srv, http.MethodDelete, "/api/v1/admin/promo-codes/NOTEXIST", nil, adminH)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("DeactivatePromoCode: status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Admin: create theme / mnemonic directly ───────────────────────────────────

func TestE2E_AdminCreateTheme_And_Mnemonic(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}

	modResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/modules",
		map[string]any{"name": "ThemeMod", "description": "D", "order_num": 10, "is_locked": false}, adminH)
	var mod map[string]any
	decodeJSON(t, modResp, &mod)
	modID := int(mod["id"].(float64))

	// Create theme
	themeResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/themes",
		map[string]any{
			"module_id": modID, "name": "ThA", "description": "d",
			"order_num": 1, "is_introduction": true, "is_locked": false,
		}, adminH)
	if themeResp.StatusCode != http.StatusCreated {
		t.Errorf("CreateTheme: status = %d, want 201", themeResp.StatusCode)
	}
	var theme map[string]any
	decodeJSON(t, themeResp, &theme)
	themeID := int(theme["id"].(float64))

	// Create text mnemonic
	mnemResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/mnemonics",
		map[string]any{"theme_id": themeID, "type": "text", "content_text": "Remember it!", "order_num": 1}, adminH)
	if mnemResp.StatusCode != http.StatusCreated {
		t.Errorf("CreateMnemonic: status = %d, want 201", mnemResp.StatusCode)
	}
	mnemResp.Body.Close()

	// Create test
	testResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/tests",
		map[string]any{
			"theme_id": themeID, "difficulty": 1, "passing_score": 70,
			"shuffle_questions": false, "shuffle_answers": false,
			"questions": []map[string]any{
				{"text": "Q?", "type": "multiple_choice", "correct_answer": "A", "order_num": 1},
			},
		}, adminH)
	if testResp.StatusCode != http.StatusCreated {
		t.Errorf("CreateTest: status = %d, want 201", testResp.StatusCode)
	}
	testResp.Body.Close()
}

// ── Teacher: bad path params ──────────────────────────────────────────────────

func TestE2E_Teacher_BadPathParams(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "7040"}

	// Bad teacher_id in GetStudents
	resp := doJSON(t, srv, http.MethodGet, "/api/v1/teachers/bad/students", nil, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("GetStudents bad id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// Bad teacher_id in GetStatistics
	resp = doJSON(t, srv, http.MethodGet, "/api/v1/teachers/bad/statistics", nil, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("GetStatistics bad id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// Bad student_id in GetStudentProgress
	resp = doJSON(t, srv, http.MethodGet, "/api/v1/teachers/7040/students/bad/progress", nil, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("GetStudentProgress bad student_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Content handler: bad request paths ────────────────────────────────────────

func TestE2E_Content_BadRequests(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "8000"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 8000, }, userH)

	// GetModules: user_id comes from auth header — valid request returns 200
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/content/modules", nil)
	req.Header.Set("X-Telegram-User-Id", "8000")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetModules with auth header: status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()

	// CreateStudySession: theme_id = 0
	resp = doJSON(t, srv, http.MethodPost, "/api/v1/users/8000/study-sessions",
		map[string]any{"theme_id": 0}, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateStudySession zero theme_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// StartTestAttempt: theme_id = 0
	resp = doJSON(t, srv, http.MethodPost, "/api/v1/users/8000/test-attempts",
		map[string]any{"theme_id": 0}, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("StartTestAttempt zero theme_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// CheckThemeAccess: bad theme_id
	resp = doJSON(t, srv, http.MethodGet, "/api/v1/users/8000/themes/bad/access", nil, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CheckThemeAccess bad theme_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// GetUserProgress: bad user_id
	resp = doJSON(t, srv, http.MethodGet, "/api/v1/users/bad/progress", nil, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("GetUserProgress bad user_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// GetModuleProgress: bad module_id
	resp = doJSON(t, srv, http.MethodGet, "/api/v1/users/8000/progress/modules/bad", nil, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("GetModuleProgress bad module_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// GetModuleThemes: user_id comes from auth header — nonexistent module returns 404
	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/api/v1/content/modules/999999/themes", nil)
	req.Header.Set("X-Telegram-User-Id", "8000")
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GetModuleThemes nonexistent module: status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Admin handler: bad request paths ─────────────────────────────────────────

func TestE2E_Admin_BadRequests(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}

	// CreatePromoCode: missing code
	resp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/promo-codes",
		map[string]any{"university_name": "U", "max_activations": 5}, adminH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreatePromoCode missing code: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// CreatePromoCode: bad max_activations
	resp = doJSON(t, srv, http.MethodPost, "/api/v1/admin/promo-codes",
		map[string]any{"code": "X", "university_name": "U", "max_activations": 0}, adminH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreatePromoCode zero max_activations: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// CreatePromoCode: bad expires_at format
	expiresStr := "not-a-date"
	resp = doJSON(t, srv, http.MethodPost, "/api/v1/admin/promo-codes",
		map[string]any{"code": "X2", "university_name": "U", "max_activations": 5, "expires_at": expiresStr}, adminH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreatePromoCode bad expires_at: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// CreateModule: invalid JSON
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/admin/content/modules",
		bytes.NewBufferString("notjson"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", testAdminToken)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateModule bad JSON: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// CreateTheme: invalid JSON
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/api/v1/admin/content/themes",
		bytes.NewBufferString("notjson"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", testAdminToken)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateTheme bad JSON: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// CreateMnemonic: invalid JSON
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/api/v1/admin/content/mnemonics",
		bytes.NewBufferString("notjson"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", testAdminToken)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateMnemonic bad JSON: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// CreateTest: invalid JSON
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/api/v1/admin/content/tests",
		bytes.NewBufferString("notjson"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", testAdminToken)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateTest bad JSON: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// CreateTest: missing theme_id
	resp = doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/tests",
		map[string]any{"difficulty": 1, "passing_score": 70,
			"questions": []map[string]any{{"text": "Q", "type": "multiple_choice", "correct_answer": "A", "order_num": 1}}},
		adminH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateTest missing theme_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// UpdateModule: bad module_id
	resp = doJSON(t, srv, http.MethodPut, "/api/v1/admin/content/modules/bad",
		map[string]any{"name": "X"}, adminH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("UpdateModule bad id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Payment handler: bad requests ─────────────────────────────────────────────

func TestE2E_Payment_BadRequests(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "8010"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 8010, }, userH)

	// CreateInvoice: bad user_id
	resp := doJSON(t, srv, http.MethodPost, "/api/v1/users/bad/payment-invoices",
		map[string]any{"plan": "monthly"}, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateInvoice bad user_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// GetPendingInvoice: bad user_id
	resp = doJSON(t, srv, http.MethodGet, "/api/v1/users/bad/payment-invoices/pending", nil, userH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("GetPendingInvoice bad user_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// CreateInvoice: invalid JSON
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/users/8010/payment-invoices",
		bytes.NewBufferString("notjson"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-User-Id", "8010")
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateInvoice bad JSON: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Admin handler: more validation paths ─────────────────────────────────────

func TestE2E_Admin_MoreValidation(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}

	// CreateTheme: missing name
	resp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/themes",
		map[string]any{"module_id": 1, "order_num": 1}, adminH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateTheme missing name: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// CreateTheme: missing module_id
	resp = doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/themes",
		map[string]any{"name": "T", "order_num": 1}, adminH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateTheme missing module_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// CreateMnemonic: missing theme_id
	resp = doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/mnemonics",
		map[string]any{"type": "text", "content_text": "X", "order_num": 1}, adminH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateMnemonic missing theme_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// CreateMnemonic: missing type
	resp = doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/mnemonics",
		map[string]any{"theme_id": 1, "order_num": 1}, adminH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateMnemonic missing type: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// UpdateModule: invalid JSON body
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/api/v1/admin/content/modules/1",
		bytes.NewBufferString("notjson"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", testAdminToken)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("UpdateModule bad JSON: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// UpdateModule: missing name
	resp = doJSON(t, srv, http.MethodPut, "/api/v1/admin/content/modules/1",
		map[string]any{"order_num": 1}, adminH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("UpdateModule missing name: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Content handler: more validation paths ────────────────────────────────────

func TestE2E_Content_MoreValidation(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "8020"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 8020, }, userH)

	// CreateStudySession: invalid JSON
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/users/8020/study-sessions",
		bytes.NewBufferString("notjson"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-User-Id", "8020")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateStudySession bad JSON: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// StartTestAttempt: invalid JSON
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/api/v1/users/8020/test-attempts",
		bytes.NewBufferString("notjson"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-User-Id", "8020")
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("StartTestAttempt bad JSON: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// SubmitTestAttempt: invalid JSON
	req, _ = http.NewRequest(http.MethodPut, srv.URL+"/api/v1/users/8020/test-attempts/some-id",
		bytes.NewBufferString("notjson"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-User-Id", "8020")
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("SubmitTestAttempt bad JSON: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// SubmitTestAttempt: bad user_id
	req, _ = http.NewRequest(http.MethodPut, srv.URL+"/api/v1/users/bad/test-attempts/some-id",
		bytes.NewBufferString(`{"answers":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-User-Id", "8020")
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("SubmitTestAttempt bad user_id: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Admin: GetUsers with filters ──────────────────────────────────────────────

func TestE2E_Admin_GetUsersFiltered(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}
	userH := map[string]string{"X-Telegram-User-Id": "9000"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 9000, }, userH)

	// GetUsers with role filter, subscription_status, valid limit and offset
	resp := doJSON(t, srv, http.MethodGet,
		"/api/v1/admin/users?role=student&subscription_status=inactive&limit=10&offset=0", nil, adminH)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetUsers filtered: status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()

	// GetUsers with valid limit (covers the strconv.Atoi success branch)
	resp = doJSON(t, srv, http.MethodGet, "/api/v1/admin/users?limit=5&offset=0", nil, adminH)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetUsers valid limit: status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Admin: CreateModule with missing name ──────────────────────────────────────

func TestE2E_Admin_CreateModule_MissingName(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}

	resp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/modules",
		map[string]any{"description": "D", "order_num": 1}, adminH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateModule missing name: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── ActivatePromoCode: missing code in body ───────────────────────────────────

func TestE2E_ActivatePromoCode_Validation(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	teacherH := map[string]string{"X-Telegram-User-Id": "9010"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 9010, }, teacherH)
	doJSON(t, srv, http.MethodPatch, "/api/v1/users/9010",
		map[string]any{"role": "teacher"}, teacherH)

	// ActivatePromoCode: invalid JSON body
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/teachers/9010/promo-codes",
		bytes.NewBufferString("notjson"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-User-Id", "9010")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("ActivatePromoCode bad JSON: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	// ActivatePromoCode: empty code
	resp = doJSON(t, srv, http.MethodPost, "/api/v1/teachers/9010/promo-codes",
		map[string]any{"code": ""}, teacherH)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("ActivatePromoCode empty code: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── UpdateUser: settings + invalid JSON ───────────────────────────────────────

func TestE2E_UpdateUser_InvalidJSON(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "9020"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 9020, }, userH)

	req, _ := http.NewRequest(http.MethodPatch, srv.URL+"/api/v1/users/9020",
		bytes.NewBufferString("notjson"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-User-Id", "9020")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("UpdateUser bad JSON: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Recovery middleware test ──────────────────────────────────────────────────

func TestE2E_RegisterUser_InvalidJSON(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/users",
		bytes.NewBufferString("notjson"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Telegram-User-Id", "9030")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("RegisterUser bad JSON: status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── Full study flow with GetModuleProgress after completion ───────────────────
// Covers: attempts.GetByUserAndTheme, test_repo.GetByID, scanAttemptRows

func TestE2E_FullStudyFlow_GetModuleProgress(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}
	userH := map[string]string{"X-Telegram-User-Id": "9100"}

	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 9100, }, userH)

	// Create module
	modResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/modules",
		map[string]any{"name": "FullMod", "description": "D", "order_num": 1, "is_locked": false}, adminH)
	var mod map[string]any
	decodeJSON(t, modResp, &mod)
	modID := int(mod["id"].(float64))

	// Create theme
	themeResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/themes",
		map[string]any{
			"module_id": modID, "name": "FullTheme", "description": "D",
			"order_num": 1, "is_introduction": true, "is_locked": false,
		}, adminH)
	var theme map[string]any
	decodeJSON(t, themeResp, &theme)
	themeID := int(theme["id"].(float64))

	// Create test
	doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/tests",
		map[string]any{
			"theme_id": themeID, "difficulty": 1, "passing_score": 50,
			"shuffle_questions": false, "shuffle_answers": false,
			"questions": []map[string]any{
				{"id": 1, "text": "Q?", "type": "multiple_choice",
					"options": []string{"A", "B"}, "correct_answer": "A", "order_num": 1},
			},
		}, adminH)

	// Study session (marks theme as started)
	doJSON(t, srv, http.MethodPost, "/api/v1/users/9100/study-sessions",
		map[string]any{"theme_id": themeID}, userH)

	// Start and submit test (marks progress)
	attemptResp := doJSON(t, srv, http.MethodPost, "/api/v1/users/9100/test-attempts",
		map[string]any{"theme_id": themeID}, userH)
	var attempt map[string]any
	decodeJSON(t, attemptResp, &attempt)
	attemptID, _ := attempt["attempt_id"].(string)

	doJSON(t, srv, http.MethodPut, "/api/v1/users/9100/test-attempts/"+attemptID,
		map[string]any{"answers": []map[string]any{{"question_id": 1, "answer": "A"}}}, userH)

	// Now get module progress — this calls attempts.GetByUserAndTheme
	progResp := doJSON(t, srv, http.MethodGet,
		"/api/v1/users/9100/progress/modules/"+itoa(modID), nil, userH)
	if progResp.StatusCode != http.StatusOK {
		t.Errorf("GetModuleProgress after completion: status = %d, want 200", progResp.StatusCode)
	}
	progResp.Body.Close()

	// GetUserProgress (also hits the completed path)
	userProgResp := doJSON(t, srv, http.MethodGet,
		"/api/v1/users/9100/progress", nil, userH)
	if userProgResp.StatusCode != http.StatusOK {
		t.Errorf("GetUserProgress after completion: status = %d, want 200", userProgResp.StatusCode)
	}
	userProgResp.Body.Close()
}

// ── GetModuleThemes with existing themes ─────────────────────────────────────
// Covers the loop body in GetModuleThemes usecase (previously 31.8%)

func TestE2E_GetModuleThemes_WithThemes(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}
	userH := map[string]string{"X-Telegram-User-Id": "9200"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 9200, }, userH)

	modResp := doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/modules",
		map[string]any{"name": "ThemeList", "description": "D", "order_num": 1, "is_locked": false}, adminH)
	var mod map[string]any
	decodeJSON(t, modResp, &mod)
	modID := int(mod["id"].(float64))

	// Create two themes
	doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/themes",
		map[string]any{
			"module_id": modID, "name": "Intro", "description": "D",
			"order_num": 1, "is_introduction": true, "is_locked": false,
		}, adminH)
	doJSON(t, srv, http.MethodPost, "/api/v1/admin/content/themes",
		map[string]any{
			"module_id": modID, "name": "Topic2", "description": "D",
			"order_num": 2, "is_introduction": false, "is_locked": false,
		}, adminH)

	// GetModuleThemes — exercises loop over themes with access check
	resp := doJSON(t, srv, http.MethodGet,
		"/api/v1/content/modules/"+itoa(modID)+"/themes?user_id=9200", nil, userH)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetModuleThemes with themes: status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── GetTeacherPromoCodes: student is not a teacher ────────────────────────────

func TestE2E_GetTeacherPromoCodes_NotTeacher(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	userH := map[string]string{"X-Telegram-User-Id": "9300"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 9300, }, userH)

	// Student tries to get promo codes — should fail with forbidden
	resp := doJSON(t, srv, http.MethodGet, "/api/v1/teachers/9300/promo-codes", nil, userH)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("GetTeacherPromoCodes not teacher: status = %d, want 403", resp.StatusCode)
	}
	resp.Body.Close()
}

// ── CreatePaymentSubscription: active sub conflict ────────────────────────────

func TestE2E_CreateSubscription_Payment_ActiveConflict(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	adminH := map[string]string{"X-Admin-Token": testAdminToken}
	userH := map[string]string{"X-Telegram-User-Id": "9310"}
	doJSON(t, srv, http.MethodPost, "/api/v1/users",
		map[string]any{"telegram_id": 9310, }, userH)

	// Create first subscription via payment
	doJSON(t, srv, http.MethodPost, "/api/v1/users/9310/subscriptions",
		map[string]any{"type": "payment", "payment_id": "pay_first", "plan": "monthly"}, userH)

	// Try second subscription via payment (different payment_id) — should conflict
	resp := doJSON(t, srv, http.MethodPost, "/api/v1/users/9310/subscriptions",
		map[string]any{"type": "payment", "payment_id": "pay_second", "plan": "monthly"}, userH)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("active sub conflict: status = %d, want 409", resp.StatusCode)
	}
	resp.Body.Close()
	_ = adminH
}

// ── helpers ───────────────────────────────────────────────────────────────────

func itoa(i int) string {
	return strconv.Itoa(i)
}
