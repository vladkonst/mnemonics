// Package admin_panel provides a web-based admin UI for the Mnemo backend.
// It is served as a separate binary on a different port (default :9000).
package admin_panel

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/internal/domain/user"
	"github.com/vladkonst/mnemonics/internal/repository/sqlite"
	adminUC "github.com/vladkonst/mnemonics/internal/usecase/admin"
)

//go:embed templates
var templateFS embed.FS

// Server holds the admin panel HTTP server and its dependencies.
type Server struct {
	db         *sql.DB
	adminToken string
	uc         *adminUC.UseCase
	tmpl       *template.Template
	router     chi.Router
}

var funcMap = template.FuncMap{
	"deref":    func(s *string) string { if s == nil { return "" }; return *s },
	"derefInt": func(i *int) int { if i == nil { return 0 }; return *i },
	"string":   func(v interface{}) string { return fmt.Sprintf("%v", v) },
	"formatTime": func(t *time.Time) string {
		if t == nil {
			return "—"
		}
		return t.Format("02.01.2006 15:04")
	},
	"len": func(v interface{}) int {
		switch val := v.(type) {
		case []content.Question:
			return len(val)
		default:
			return 0
		}
	},
}

// NewServer creates a new admin panel server.
func NewServer(db *sql.DB, adminToken string) (*Server, error) {
	userRepo := sqlite.NewUserRepo(db)
	moduleRepo := sqlite.NewModuleRepo(db)
	themeRepo := sqlite.NewThemeRepo(db)
	mnemonicRepo := sqlite.NewMnemonicRepo(db)
	testRepo := sqlite.NewTestRepo(db)
	promoCodeRepo := sqlite.NewPromoCodeRepo(db)

	uc := adminUC.NewUseCase(moduleRepo, themeRepo, mnemonicRepo, testRepo, promoCodeRepo, userRepo, db)

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS,
		"templates/base.html",
		"templates/login.html",
		"templates/dashboard.html",
		"templates/modules/list.html",
		"templates/modules/form.html",
		"templates/themes/list.html",
		"templates/themes/form.html",
		"templates/mnemonics/list.html",
		"templates/mnemonics/form.html",
		"templates/tests/list.html",
		"templates/tests/form.html",
		"templates/promo_codes/list.html",
		"templates/promo_codes/form.html",
		"templates/users/list.html",
	)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	s := &Server{db: db, adminToken: adminToken, uc: uc, tmpl: tmpl}
	s.buildRouter()
	return s, nil
}

// Handler returns the HTTP handler for the admin panel.
func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) buildRouter() {
	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)

	r.Get("/admin/login", s.loginPage)
	r.Post("/admin/login", s.loginSubmit)
	r.Get("/admin/logout", s.logout)

	r.Group(func(r chi.Router) {
		r.Use(s.authMiddleware)

		r.Get("/admin/", s.dashboard)
		r.Get("/admin", s.dashboard)

		r.Get("/admin/modules", s.modulesList)
		r.Get("/admin/modules/new", s.modulesNew)
		r.Post("/admin/modules/new", s.modulesCreate)
		r.Get("/admin/modules/{id}/edit", s.modulesEdit)
		r.Post("/admin/modules/{id}/edit", s.modulesUpdate)
		r.Post("/admin/modules/{id}/delete", s.modulesDelete)

		r.Get("/admin/themes", s.themesList)
		r.Get("/admin/themes/new", s.themesNew)
		r.Post("/admin/themes/new", s.themesCreate)
		r.Get("/admin/themes/{id}/edit", s.themesEdit)
		r.Post("/admin/themes/{id}/edit", s.themesUpdate)
		r.Post("/admin/themes/{id}/delete", s.themesDelete)

		r.Get("/admin/mnemonics", s.mnemonicsList)
		r.Get("/admin/mnemonics/new", s.mnemonicsNew)
		r.Post("/admin/mnemonics/new", s.mnemonicsCreate)
		r.Get("/admin/mnemonics/{id}/edit", s.mnemonicsEdit)
		r.Post("/admin/mnemonics/{id}/edit", s.mnemonicsUpdate)
		r.Post("/admin/mnemonics/{id}/delete", s.mnemonicsDelete)

		r.Get("/admin/tests", s.testsList)
		r.Get("/admin/tests/new", s.testsNew)
		r.Post("/admin/tests/new", s.testsCreate)
		r.Get("/admin/tests/{id}/edit", s.testsEdit)
		r.Post("/admin/tests/{id}/edit", s.testsUpdate)
		r.Post("/admin/tests/{id}/delete", s.testsDelete)

		r.Get("/admin/promo-codes", s.promoList)
		r.Get("/admin/promo-codes/new", s.promoNew)
		r.Post("/admin/promo-codes/new", s.promoCreate)
		r.Post("/admin/promo-codes/{code}/deactivate", s.promoDeactivate)

		r.Get("/admin/users", s.usersList)
	})

	s.router = r
}

// ── Auth ──────────────────────────────────────────────────────────────────────

const sessionCookie = "admin_session"

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookie)
		if err != nil || !s.validSession(r.Context(), c.Value) {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) validSession(ctx context.Context, token string) bool {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM admin_sessions WHERE token = ?", token).Scan(&count)
	return err == nil && count > 0
}

func (s *Server) loginPage(w http.ResponseWriter, r *http.Request) {
	s.render(w, "login.html", map[string]interface{}{"Error": r.URL.Query().Get("err")})
}

func (s *Server) loginSubmit(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if token != s.adminToken {
		http.Redirect(w, r, "/admin/login?err=Неверный+токен", http.StatusFound)
		return
	}
	sessionToken := fmt.Sprintf("%d-%s", time.Now().UnixNano(), safePrefix(token, 8))
	_, _ = s.db.ExecContext(r.Context(), "INSERT OR IGNORE INTO admin_sessions (token) VALUES (?)", sessionToken)
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: sessionToken,
		Path: "/admin", HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/admin/", http.StatusFound)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		_, _ = s.db.ExecContext(r.Context(), "DELETE FROM admin_sessions WHERE token = ?", c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", MaxAge: -1, Path: "/admin"})
	http.Redirect(w, r, "/admin/login", http.StatusFound)
}

// ── Dashboard ─────────────────────────────────────────────────────────────────

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := s.uc.GetAnalytics(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "dashboard.html", map[string]interface{}{"Active": "dashboard", "Stats": stats})
}

// ── Modules ───────────────────────────────────────────────────────────────────

func (s *Server) modulesList(w http.ResponseWriter, r *http.Request) {
	modules, err := s.uc.GetModules(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "modules/list.html", map[string]interface{}{"Active": "modules", "Modules": modules})
}

func (s *Server) modulesNew(w http.ResponseWriter, r *http.Request) {
	s.render(w, "modules/form.html", map[string]interface{}{"Active": "modules"})
}

func (s *Server) modulesCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		s.renderFlash(w, "modules/form.html", map[string]interface{}{"Active": "modules"}, "Название обязательно")
		return
	}
	orderNum, _ := strconv.Atoi(r.FormValue("order_num"))
	desc := strings.TrimSpace(r.FormValue("description"))
	emoji := strings.TrimSpace(r.FormValue("icon_emoji"))
	isLocked := r.FormValue("is_locked") == "1"
	var emojiPtr *string
	if emoji != "" {
		emojiPtr = &emoji
	}
	if _, err := s.uc.CreateModule(r.Context(), name, desc, orderNum, isLocked, emojiPtr); err != nil {
		s.renderFlash(w, "modules/form.html", map[string]interface{}{"Active": "modules"}, err.Error())
		return
	}
	http.Redirect(w, r, "/admin/modules", http.StatusFound)
}

func (s *Server) modulesEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	module, err := s.uc.GetModuleByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Модуль не найден", http.StatusNotFound)
		return
	}
	s.render(w, "modules/form.html", map[string]interface{}{"Active": "modules", "Module": module})
}

func (s *Server) modulesUpdate(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	orderNum, _ := strconv.Atoi(r.FormValue("order_num"))
	desc := strings.TrimSpace(r.FormValue("description"))
	emoji := strings.TrimSpace(r.FormValue("icon_emoji"))
	isLocked := r.FormValue("is_locked") == "1"
	var emojiPtr *string
	if emoji != "" {
		emojiPtr = &emoji
	}
	if _, err := s.uc.UpdateModule(r.Context(), id, name, desc, orderNum, isLocked, emojiPtr); err != nil {
		s.renderFlash(w, "modules/form.html", map[string]interface{}{"Active": "modules"}, err.Error())
		return
	}
	http.Redirect(w, r, "/admin/modules", http.StatusFound)
}

func (s *Server) modulesDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := s.uc.DeleteModule(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/modules", http.StatusFound)
}

// ── Themes ────────────────────────────────────────────────────────────────────

func (s *Server) themesList(w http.ResponseWriter, r *http.Request) {
	themes, err := s.uc.GetAllThemes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "themes/list.html", map[string]interface{}{"Active": "themes", "Themes": themes})
}

func (s *Server) themesNew(w http.ResponseWriter, r *http.Request) {
	modules, _ := s.uc.GetModules(r.Context())
	s.render(w, "themes/form.html", map[string]interface{}{"Active": "themes", "Modules": modules})
}

func (s *Server) themesCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	moduleID, _ := strconv.Atoi(r.FormValue("module_id"))
	name := strings.TrimSpace(r.FormValue("name"))
	desc := strings.TrimSpace(r.FormValue("description"))
	orderNum, _ := strconv.Atoi(r.FormValue("order_num"))
	isIntro := r.FormValue("is_introduction") == "1"
	isLocked := r.FormValue("is_locked") == "1"
	estMins := parseOptionalInt(r.FormValue("estimated_time_minutes"))
	if _, err := s.uc.CreateTheme(r.Context(), moduleID, name, desc, orderNum, isIntro, isLocked, estMins); err != nil {
		modules, _ := s.uc.GetModules(r.Context())
		s.renderFlash(w, "themes/form.html", map[string]interface{}{"Active": "themes", "Modules": modules}, err.Error())
		return
	}
	http.Redirect(w, r, "/admin/themes", http.StatusFound)
}

func (s *Server) themesEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	theme, err := s.uc.GetThemeByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Тема не найдена", http.StatusNotFound)
		return
	}
	modules, _ := s.uc.GetModules(r.Context())
	s.render(w, "themes/form.html", map[string]interface{}{"Active": "themes", "Theme": theme, "Modules": modules})
}

func (s *Server) themesUpdate(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	desc := strings.TrimSpace(r.FormValue("description"))
	var descPtr *string
	if desc != "" {
		descPtr = &desc
	}
	orderNum, _ := strconv.Atoi(r.FormValue("order_num"))
	isLocked := r.FormValue("is_locked") == "1"
	estMins := parseOptionalInt(r.FormValue("estimated_time_minutes"))
	if _, err := s.uc.UpdateTheme(r.Context(), id, name, descPtr, orderNum, isLocked, estMins); err != nil {
		modules, _ := s.uc.GetModules(r.Context())
		s.renderFlash(w, "themes/form.html", map[string]interface{}{"Active": "themes", "Modules": modules}, err.Error())
		return
	}
	http.Redirect(w, r, "/admin/themes", http.StatusFound)
}

func (s *Server) themesDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := s.uc.DeleteTheme(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/themes", http.StatusFound)
}

// ── Mnemonics ─────────────────────────────────────────────────────────────────

func (s *Server) mnemonicsList(w http.ResponseWriter, r *http.Request) {
	mnemonics, err := s.uc.GetAllMnemonics(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "mnemonics/list.html", map[string]interface{}{"Active": "mnemonics", "Mnemonics": mnemonics})
}

func (s *Server) mnemonicsNew(w http.ResponseWriter, r *http.Request) {
	themes, _ := s.uc.GetAllThemes(r.Context())
	s.render(w, "mnemonics/form.html", map[string]interface{}{"Active": "mnemonics", "Themes": themes})
}

func (s *Server) mnemonicsCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	themeID, _ := strconv.Atoi(r.FormValue("theme_id"))
	typ := content.MnemonicType(r.FormValue("type"))
	orderNum, _ := strconv.Atoi(r.FormValue("order_num"))
	var textPtr, s3Ptr *string
	if t := r.FormValue("content_text"); t != "" {
		textPtr = &t
	}
	if s3 := r.FormValue("s3_image_key"); s3 != "" {
		s3Ptr = &s3
	}
	if _, err := s.uc.CreateMnemonic(r.Context(), themeID, typ, textPtr, s3Ptr, orderNum); err != nil {
		themes, _ := s.uc.GetAllThemes(r.Context())
		s.renderFlash(w, "mnemonics/form.html", map[string]interface{}{"Active": "mnemonics", "Themes": themes}, err.Error())
		return
	}
	http.Redirect(w, r, "/admin/mnemonics", http.StatusFound)
}

func (s *Server) mnemonicsEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	mn, err := s.uc.GetMnemonicByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Мнемоника не найдена", http.StatusNotFound)
		return
	}
	themes, _ := s.uc.GetAllThemes(r.Context())
	s.render(w, "mnemonics/form.html", map[string]interface{}{"Active": "mnemonics", "Mnemonic": mn, "Themes": themes})
}

func (s *Server) mnemonicsUpdate(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	orderNum, _ := strconv.Atoi(r.FormValue("order_num"))
	var textPtr, s3Ptr *string
	if t := r.FormValue("content_text"); t != "" {
		textPtr = &t
	}
	if s3 := r.FormValue("s3_image_key"); s3 != "" {
		s3Ptr = &s3
	}
	if _, err := s.uc.UpdateMnemonic(r.Context(), id, textPtr, s3Ptr, orderNum); err != nil {
		themes, _ := s.uc.GetAllThemes(r.Context())
		s.renderFlash(w, "mnemonics/form.html", map[string]interface{}{"Active": "mnemonics", "Themes": themes}, err.Error())
		return
	}
	http.Redirect(w, r, "/admin/mnemonics", http.StatusFound)
}

func (s *Server) mnemonicsDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := s.uc.DeleteMnemonic(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/mnemonics", http.StatusFound)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func (s *Server) testsList(w http.ResponseWriter, r *http.Request) {
	tests, err := s.uc.GetAllTests(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "tests/list.html", map[string]interface{}{"Active": "tests", "Tests": tests})
}

func (s *Server) testsNew(w http.ResponseWriter, r *http.Request) {
	themes, _ := s.uc.GetAllThemes(r.Context())
	s.render(w, "tests/form.html", map[string]interface{}{"Active": "tests", "Themes": themes, "QuestionsJSON": "[]"})
}

func (s *Server) testsCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	themeID, _ := strconv.Atoi(r.FormValue("theme_id"))
	difficulty, _ := strconv.Atoi(r.FormValue("difficulty"))
	passingScore, _ := strconv.Atoi(r.FormValue("passing_score"))
	shuffleQ := r.FormValue("shuffle_questions") == "1"
	shuffleA := r.FormValue("shuffle_answers") == "1"
	questions, err := parseQuestionsJSON(r.FormValue("questions_json"))
	if err != nil {
		themes, _ := s.uc.GetAllThemes(r.Context())
		s.renderFlash(w, "tests/form.html", map[string]interface{}{"Active": "tests", "Themes": themes, "QuestionsJSON": r.FormValue("questions_json")}, "Ошибка JSON: "+err.Error())
		return
	}
	if _, err := s.uc.CreateTest(r.Context(), themeID, difficulty, passingScore, shuffleQ, shuffleA, questions); err != nil {
		themes, _ := s.uc.GetAllThemes(r.Context())
		s.renderFlash(w, "tests/form.html", map[string]interface{}{"Active": "tests", "Themes": themes, "QuestionsJSON": r.FormValue("questions_json")}, err.Error())
		return
	}
	http.Redirect(w, r, "/admin/tests", http.StatusFound)
}

func (s *Server) testsEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	test, err := s.uc.GetTestByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Тест не найден", http.StatusNotFound)
		return
	}
	qJSON, _ := json.MarshalIndent(test.Questions, "", "  ")
	themes, _ := s.uc.GetAllThemes(r.Context())
	s.render(w, "tests/form.html", map[string]interface{}{
		"Active": "tests", "Test": test, "Themes": themes, "QuestionsJSON": string(qJSON),
	})
}

func (s *Server) testsUpdate(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	difficulty, _ := strconv.Atoi(r.FormValue("difficulty"))
	passingScore, _ := strconv.Atoi(r.FormValue("passing_score"))
	shuffleQ := r.FormValue("shuffle_questions") == "1"
	shuffleA := r.FormValue("shuffle_answers") == "1"
	questions, err := parseQuestionsJSON(r.FormValue("questions_json"))
	if err != nil {
		themes, _ := s.uc.GetAllThemes(r.Context())
		s.renderFlash(w, "tests/form.html", map[string]interface{}{"Active": "tests", "Themes": themes, "QuestionsJSON": r.FormValue("questions_json")}, "Ошибка JSON: "+err.Error())
		return
	}
	if _, err := s.uc.UpdateTest(r.Context(), id, difficulty, passingScore, shuffleQ, shuffleA, questions); err != nil {
		themes, _ := s.uc.GetAllThemes(r.Context())
		s.renderFlash(w, "tests/form.html", map[string]interface{}{"Active": "tests", "Themes": themes, "QuestionsJSON": r.FormValue("questions_json")}, err.Error())
		return
	}
	http.Redirect(w, r, "/admin/tests", http.StatusFound)
}

func (s *Server) testsDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := s.uc.DeleteTest(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/tests", http.StatusFound)
}

// ── Promo Codes ───────────────────────────────────────────────────────────────

func (s *Server) promoList(w http.ResponseWriter, r *http.Request) {
	codes, err := s.uc.GetAllPromoCodes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "promo_codes/list.html", map[string]interface{}{"Active": "promo_codes", "PromoCodes": codes})
}

func (s *Server) promoNew(w http.ResponseWriter, r *http.Request) {
	s.render(w, "promo_codes/form.html", map[string]interface{}{"Active": "promo_codes"})
}

func (s *Server) promoCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	code := strings.TrimSpace(r.FormValue("code"))
	univName := strings.TrimSpace(r.FormValue("university_name"))
	maxAct, _ := strconv.Atoi(r.FormValue("max_activations"))
	var expiresAt *time.Time
	if raw := r.FormValue("expires_at"); raw != "" {
		t, err := time.Parse("2006-01-02T15:04", raw)
		if err == nil {
			expiresAt = &t
		}
	}
	if _, err := s.uc.CreatePromoCode(r.Context(), code, univName, maxAct, expiresAt); err != nil {
		s.renderFlash(w, "promo_codes/form.html", map[string]interface{}{"Active": "promo_codes"}, err.Error())
		return
	}
	http.Redirect(w, r, "/admin/promo-codes", http.StatusFound)
}

func (s *Server) promoDeactivate(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if err := s.uc.DeactivatePromoCode(r.Context(), code); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/promo-codes", http.StatusFound)
}

// ── Users ─────────────────────────────────────────────────────────────────────

func (s *Server) usersList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	roleStr := q.Get("role")
	subStatus := q.Get("subscription_status")

	var rolePtr *user.Role
	if roleStr != "" {
		rv := user.Role(roleStr)
		rolePtr = &rv
	}
	var subPtr *user.SubscriptionStatus
	if subStatus != "" {
		sv := user.SubscriptionStatus(subStatus)
		subPtr = &sv
	}
	users, total, err := s.uc.GetUsers(r.Context(), rolePtr, subPtr, 200, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "users/list.html", map[string]interface{}{
		"Active": "users",
		"Users":  users,
		"Total":  total,
		"Filter": map[string]string{"Role": roleStr, "SubStatus": subStatus},
	})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (s *Server) render(w http.ResponseWriter, name string, data map[string]interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) renderFlash(w http.ResponseWriter, name string, data map[string]interface{}, msg string) {
	data["Flash"] = msg
	data["FlashType"] = "danger"
	s.render(w, name, data)
}

func parseOptionalInt(s string) *int {
	if s == "" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return nil
	}
	return &v
}

func parseQuestionsJSON(s string) ([]content.Question, error) {
	var questions []content.Question
	if err := json.Unmarshal([]byte(s), &questions); err != nil {
		return nil, err
	}
	return questions, nil
}

func safePrefix(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}
