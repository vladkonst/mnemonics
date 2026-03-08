package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vladkonst/mnemonics/internal/delivery/http/respond"
	adminUC "github.com/vladkonst/mnemonics/internal/usecase/admin"
	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/internal/domain/user"
)

// AdminHandler handles admin HTTP endpoints.
type AdminHandler struct {
	uc *adminUC.UseCase
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(uc *adminUC.UseCase) *AdminHandler {
	return &AdminHandler{uc: uc}
}

// createPromoCodeRequest is the JSON body for POST /api/v1/admin/promo-codes.
type createPromoCodeRequest struct {
	Code           string  `json:"code"`
	UniversityName string  `json:"university_name"`
	MaxActivations int     `json:"max_activations"`
	ExpiresAt      *string `json:"expires_at"` // RFC3339
}

// CreatePromoCode handles POST /api/v1/admin/promo-codes.
func (h *AdminHandler) CreatePromoCode(w http.ResponseWriter, r *http.Request) {
	var req createPromoCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.Code == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "code is required")
		return
	}
	if req.MaxActivations <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "max_activations must be positive")
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, "bad_request", "expires_at must be RFC3339 format")
			return
		}
		expiresAt = &t
	}

	promo, err := h.uc.CreatePromoCode(r.Context(), req.Code, req.UniversityName, req.MaxActivations, expiresAt)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/admin/promo-codes/%s", promo.Code))
	respond.JSON(w, http.StatusCreated, promo)
}

// DeactivatePromoCode handles DELETE /api/v1/admin/promo-codes/{code}.
func (h *AdminHandler) DeactivatePromoCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "code is required")
		return
	}

	if err := h.uc.DeactivatePromoCode(r.Context(), code); err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// createModuleRequest is the JSON body for POST /api/v1/admin/content/modules.
type createModuleRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	OrderNum    int     `json:"order_num"`
	IsLocked    bool    `json:"is_locked"`
	IconEmoji   *string `json:"icon_emoji"`
}

// CreateModule handles POST /api/v1/admin/content/modules.
func (h *AdminHandler) CreateModule(w http.ResponseWriter, r *http.Request) {
	var req createModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "name is required")
		return
	}

	module, err := h.uc.CreateModule(r.Context(), req.Name, req.Description, req.OrderNum, req.IsLocked, req.IconEmoji)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/admin/content/modules/%d", module.ID))
	respond.JSON(w, http.StatusCreated, module)
}

// UpdateModule handles PUT /api/v1/admin/content/modules/{id}.
func (h *AdminHandler) UpdateModule(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}

	var req createModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "name is required")
		return
	}

	module, err := h.uc.UpdateModule(r.Context(), id, req.Name, req.Description, req.OrderNum, req.IsLocked, req.IconEmoji)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, module)
}

// createThemeRequest is the JSON body for POST /api/v1/admin/content/themes.
type createThemeRequest struct {
	ModuleID             int     `json:"module_id"`
	Name                 string  `json:"name"`
	Description          string  `json:"description"`
	OrderNum             int     `json:"order_num"`
	IsIntroduction       bool    `json:"is_introduction"`
	IsLocked             bool    `json:"is_locked"`
	EstimatedTimeMinutes *int    `json:"estimated_time_minutes"`
}

// CreateTheme handles POST /api/v1/admin/content/themes.
func (h *AdminHandler) CreateTheme(w http.ResponseWriter, r *http.Request) {
	var req createThemeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "name is required")
		return
	}
	if req.ModuleID <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "module_id is required")
		return
	}

	theme, err := h.uc.CreateTheme(r.Context(), req.ModuleID, req.Name, req.Description, req.OrderNum, req.IsIntroduction, req.IsLocked, req.EstimatedTimeMinutes)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/admin/content/themes/%d", theme.ID))
	respond.JSON(w, http.StatusCreated, theme)
}

// createMnemonicRequest is the JSON body for POST /api/v1/admin/content/mnemonics.
type createMnemonicRequest struct {
	ThemeID     int     `json:"theme_id"`
	Type        string  `json:"type"`
	ContentText *string `json:"content_text"`
	S3ImageKey  *string `json:"s3_image_key"`
	OrderNum    int     `json:"order_num"`
}

// CreateMnemonic handles POST /api/v1/admin/content/mnemonics.
func (h *AdminHandler) CreateMnemonic(w http.ResponseWriter, r *http.Request) {
	var req createMnemonicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.ThemeID <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "theme_id is required")
		return
	}
	if req.Type == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "type is required")
		return
	}

	mnemonic, err := h.uc.CreateMnemonic(r.Context(), req.ThemeID, content.MnemonicType(req.Type), req.ContentText, req.S3ImageKey, req.OrderNum)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/admin/content/mnemonics/%d", mnemonic.ID))
	respond.JSON(w, http.StatusCreated, mnemonic)
}

// createTestRequest is the JSON body for POST /api/v1/admin/content/tests.
type createTestRequest struct {
	ThemeID          int               `json:"theme_id"`
	Difficulty       int               `json:"difficulty"`
	PassingScore     int               `json:"passing_score"`
	ShuffleQuestions bool              `json:"shuffle_questions"`
	ShuffleAnswers   bool              `json:"shuffle_answers"`
	Questions        []questionRequest `json:"questions"`
}

type questionRequest struct {
	ID            int      `json:"id"`
	Text          string   `json:"text"`
	Type          string   `json:"type"`
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correct_answer"`
	OrderNum      int      `json:"order_num"`
}

// CreateTest handles POST /api/v1/admin/content/tests.
func (h *AdminHandler) CreateTest(w http.ResponseWriter, r *http.Request) {
	var req createTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.ThemeID <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "theme_id is required")
		return
	}

	questions := make([]content.Question, 0, len(req.Questions))
	for _, q := range req.Questions {
		questions = append(questions, content.Question{
			ID:            q.ID,
			Text:          q.Text,
			Type:          content.QuestionType(q.Type),
			Options:       q.Options,
			CorrectAnswer: q.CorrectAnswer,
			OrderNum:      q.OrderNum,
		})
	}

	test, err := h.uc.CreateTest(r.Context(), req.ThemeID, req.Difficulty, req.PassingScore, req.ShuffleQuestions, req.ShuffleAnswers, questions)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/admin/content/tests/%d", test.ID))
	respond.JSON(w, http.StatusCreated, test)
}

// GetUsers handles GET /api/v1/admin/users.
func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var rolePtr *user.Role
	if roleStr := q.Get("role"); roleStr != "" {
		r := user.Role(roleStr)
		rolePtr = &r
	}

	var subStatusPtr *user.SubscriptionStatus
	if ssStr := q.Get("subscription_status"); ssStr != "" {
		ss := user.SubscriptionStatus(ssStr)
		subStatusPtr = &ss
	}

	limit := 50
	offset := 0
	if lStr := q.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	if oStr := q.Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	users, total, err := h.uc.GetUsers(r.Context(), rolePtr, subStatusPtr, limit, offset)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
		"total": total,
	})
}

// GetAnalytics handles GET /api/v1/admin/analytics/overview.
// Returns an empty stub for now.
func (h *AdminHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"analytics": map[string]interface{}{},
	})
}
