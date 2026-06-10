package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/vladkonst/mnemonics/internal/delivery/http/respond"
	"github.com/vladkonst/mnemonics/internal/domain/content"
	feedbackDomain "github.com/vladkonst/mnemonics/internal/domain/feedback"
	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
	"github.com/vladkonst/mnemonics/internal/domain/user"
	adminUC "github.com/vladkonst/mnemonics/internal/usecase/admin"
	managerUC "github.com/vladkonst/mnemonics/internal/usecase/manager"
)

// AdminHandler handles admin HTTP endpoints.
type AdminHandler struct {
	uc         *adminUC.UseCase
	managerUC  *managerUC.UseCase
	storage    interfaces.StorageService
	uploadsDir string
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(uc *adminUC.UseCase, managerUseCase *managerUC.UseCase, storage interfaces.StorageService, uploadsDir string) *AdminHandler {
	return &AdminHandler{uc: uc, managerUC: managerUseCase, storage: storage, uploadsDir: uploadsDir}
}

// UploadImage handles POST /api/v1/admin/upload.
// Uploads the file to storage and returns the object key.
func (h *AdminHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "failed to parse form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "file is required")
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	key := uuid.New().String() + ext
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if err := h.storage.UploadFile(r.Context(), key, file, header.Size, contentType); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal", "failed to upload file")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"key": key})
}

// ServeUpload handles GET /api/v1/uploads/{filename}.
// Serves the file locally if available, otherwise redirects to a pre-signed storage URL.
func (h *AdminHandler) ServeUpload(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid filename")
		return
	}

	localPath := filepath.Join(h.uploadsDir, filename)
	if _, err := os.Stat(localPath); err == nil {
		http.ServeFile(w, r, localPath)
		return
	}

	url, err := h.storage.PresignURL(r.Context(), filename)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "not_found", "file not found")
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
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
	TermRu      *string `json:"term_ru"`
	TermLatin   *string `json:"term_latin"`
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

	mnemonic, err := h.uc.CreateMnemonic(r.Context(), req.ThemeID, content.MnemonicType(req.Type), req.ContentText, req.S3ImageKey, req.TermRu, req.TermLatin, req.OrderNum)
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
	Questions        []questionRequest `json:"questions"`
}

type questionRequest struct {
	ID            int    `json:"id"`
	Text          string `json:"text"`
	CorrectAnswer string `json:"correct_answer"`
	OrderNum      int    `json:"order_num"`
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
			CorrectAnswer: q.CorrectAnswer,
			OrderNum:      q.OrderNum,
		})
	}

	test, err := h.uc.CreateTest(r.Context(), req.ThemeID, req.Difficulty, req.PassingScore, req.ShuffleQuestions, questions)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/admin/content/tests/%d", test.ID))
	respond.JSON(w, http.StatusCreated, test)
}

// createModuleTestRequest is the JSON body for POST /api/v1/admin/content/module-tests.
type createModuleTestRequest struct {
	ModuleID         int               `json:"module_id"`
	Name             string            `json:"name"`
	Difficulty       int               `json:"difficulty"`
	PassingScore     int               `json:"passing_score"`
	ShuffleQuestions bool              `json:"shuffle_questions"`
	Questions        []questionRequest `json:"questions"`
}

// generateModuleTestRequest is the JSON body for POST /api/v1/admin/content/module-tests/generate.
type generateModuleTestRequest struct {
	ModuleID         int    `json:"module_id"`
	// Deprecated fields kept for backwards compatibility; values are ignored.
	Name             string `json:"name"`
	Pct              int    `json:"pct"`
	Difficulty       int    `json:"difficulty"`
	PassingScore     int    `json:"passing_score"`
	ShuffleQuestions bool   `json:"shuffle_questions"`
}

// CreateModuleTest handles POST /api/v1/admin/content/module-tests.
func (h *AdminHandler) CreateModuleTest(w http.ResponseWriter, r *http.Request) {
	var req createModuleTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.ModuleID <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "module_id is required")
		return
	}
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "name is required")
		return
	}

	questions := make([]content.Question, 0, len(req.Questions))
	for _, q := range req.Questions {
		questions = append(questions, content.Question{
			ID:            q.ID,
			Text:          q.Text,
			CorrectAnswer: q.CorrectAnswer,
			OrderNum:      q.OrderNum,
		})
	}

	test, err := h.uc.CreateModuleTest(r.Context(), req.ModuleID, req.Name, req.Difficulty, req.PassingScore, req.ShuffleQuestions, questions)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/admin/content/module-tests/%d", test.ID))
	respond.JSON(w, http.StatusCreated, test)
}

// GenerateModuleTest handles POST /api/v1/admin/content/module-tests/generate.
func (h *AdminHandler) GenerateModuleTest(w http.ResponseWriter, r *http.Request) {
	var req generateModuleTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.ModuleID <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "module_id is required")
		return
	}
	test, err := h.uc.GenerateModuleTest(r.Context(), req.ModuleID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/admin/content/module-tests/%d", test.ID))
	respond.JSON(w, http.StatusCreated, test)
}

// GetAdminModuleTests handles GET /api/v1/admin/content/module-tests[?module_id={id}].
func (h *AdminHandler) GetAdminModuleTests(w http.ResponseWriter, r *http.Request) {
	if moduleIDStr := r.URL.Query().Get("module_id"); moduleIDStr != "" {
		moduleID, err := strconv.Atoi(moduleIDStr)
		if err != nil || moduleID <= 0 {
			respond.Error(w, http.StatusBadRequest, "bad_request", "module_id must be a valid positive integer")
			return
		}
		tests, err := h.uc.GetModuleTestsByModule(r.Context(), moduleID)
		if err != nil {
			respond.ErrorFrom(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, map[string]interface{}{"data": tests, "total": len(tests)})
		return
	}
	tests, err := h.uc.GetAllModuleTests(r.Context())
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]interface{}{"data": tests, "total": len(tests)})
}

// GetAdminModuleTest handles GET /api/v1/admin/content/module-tests/{id}.
func (h *AdminHandler) GetAdminModuleTest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}
	t, err := h.uc.GetModuleTestByID(r.Context(), id)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, t)
}

// UpdateModuleTest handles PUT /api/v1/admin/content/module-tests/{id}.
func (h *AdminHandler) UpdateModuleTest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}
	var req createModuleTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "name is required")
		return
	}
	questions := make([]content.Question, 0, len(req.Questions))
	for _, q := range req.Questions {
		questions = append(questions, content.Question{
			ID:            q.ID,
			Text:          q.Text,
			CorrectAnswer: q.CorrectAnswer,
			OrderNum:      q.OrderNum,
		})
	}
	t, err := h.uc.UpdateModuleTest(r.Context(), id, req.Name, req.Difficulty, req.PassingScore, req.ShuffleQuestions, questions)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, t)
}

// DeleteModuleTest handles DELETE /api/v1/admin/content/module-tests/{id}.
func (h *AdminHandler) DeleteModuleTest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}
	if err := h.uc.DeleteModuleTest(r.Context(), id); err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
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
func (h *AdminHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	result, err := h.uc.GetAnalytics(r.Context())
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, result)
}

// DeleteModule handles DELETE /api/v1/admin/content/modules/{id}.
func (h *AdminHandler) DeleteModule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}
	if err := h.uc.DeleteModule(r.Context(), id); err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

type updateThemeRequest struct {
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	OrderNum      int     `json:"order_num"`
	IsLocked      bool    `json:"is_locked"`
	EstimatedMins *int    `json:"estimated_time_minutes"`
}

// UpdateTheme handles PUT /api/v1/admin/content/themes/{id}.
func (h *AdminHandler) UpdateTheme(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}
	var req updateThemeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "name is required")
		return
	}
	theme, err := h.uc.UpdateTheme(r.Context(), id, req.Name, req.Description, req.OrderNum, req.IsLocked, req.EstimatedMins)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, theme)
}

// DeleteTheme handles DELETE /api/v1/admin/content/themes/{id}.
func (h *AdminHandler) DeleteTheme(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}
	if err := h.uc.DeleteTheme(r.Context(), id); err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

type updateMnemonicRequest struct {
	ContentText *string `json:"content_text"`
	S3ImageKey  *string `json:"s3_image_key"`
	TermRu      *string `json:"term_ru"`
	TermLatin   *string `json:"term_latin"`
	OrderNum    int     `json:"order_num"`
}

// UpdateMnemonic handles PUT /api/v1/admin/content/mnemonics/{id}.
func (h *AdminHandler) UpdateMnemonic(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}
	var req updateMnemonicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	mnemonic, err := h.uc.UpdateMnemonic(r.Context(), id, req.ContentText, req.S3ImageKey, req.TermRu, req.TermLatin, req.OrderNum)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, mnemonic)
}

// DeleteMnemonic handles DELETE /api/v1/admin/content/mnemonics/{id}.
func (h *AdminHandler) DeleteMnemonic(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}
	if err := h.uc.DeleteMnemonic(r.Context(), id); err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// UpdateTest handles PUT /api/v1/admin/content/tests/{id}.
func (h *AdminHandler) UpdateTest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}
	var req createTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	questions := make([]content.Question, 0, len(req.Questions))
	for _, q := range req.Questions {
		questions = append(questions, content.Question{
			ID:            q.ID,
			Text:          q.Text,
			CorrectAnswer: q.CorrectAnswer,
			OrderNum:      q.OrderNum,
		})
	}
	test, err := h.uc.UpdateTest(r.Context(), id, req.Difficulty, req.PassingScore, req.ShuffleQuestions, questions)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, test)
}

// DeleteTest handles DELETE /api/v1/admin/content/tests/{id}.
func (h *AdminHandler) DeleteTest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}
	if err := h.uc.DeleteTest(r.Context(), id); err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// GetAdminModules handles GET /api/v1/admin/content/modules.
func (h *AdminHandler) GetAdminModules(w http.ResponseWriter, r *http.Request) {
	modules, err := h.uc.GetModules(r.Context())
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]interface{}{"data": modules, "total": len(modules)})
}

// GetAdminModule handles GET /api/v1/admin/content/modules/{id}.
func (h *AdminHandler) GetAdminModule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}
	m, err := h.uc.GetModuleByID(r.Context(), id)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, m)
}

// GetAdminThemes handles GET /api/v1/admin/content/themes.
func (h *AdminHandler) GetAdminThemes(w http.ResponseWriter, r *http.Request) {
	themes, err := h.uc.GetAllThemes(r.Context())
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]interface{}{"data": themes, "total": len(themes)})
}

// GetAdminTheme handles GET /api/v1/admin/content/themes/{id}.
func (h *AdminHandler) GetAdminTheme(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}
	t, err := h.uc.GetThemeByID(r.Context(), id)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, t)
}

// GetAdminMnemonics handles GET /api/v1/admin/content/mnemonics.
func (h *AdminHandler) GetAdminMnemonics(w http.ResponseWriter, r *http.Request) {
	mnemonics, err := h.uc.GetAllMnemonics(r.Context())
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]interface{}{"data": mnemonics, "total": len(mnemonics)})
}

// GetAdminMnemonic handles GET /api/v1/admin/content/mnemonics/{id}.
func (h *AdminHandler) GetAdminMnemonic(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}
	m, err := h.uc.GetMnemonicByID(r.Context(), id)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, m)
}

// GetAdminTests handles GET /api/v1/admin/content/tests.
func (h *AdminHandler) GetAdminTests(w http.ResponseWriter, r *http.Request) {
	tests, err := h.uc.GetAllTests(r.Context())
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]interface{}{"data": tests, "total": len(tests)})
}

// GetAdminTest handles GET /api/v1/admin/content/tests/{id}.
func (h *AdminHandler) GetAdminTest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id must be a valid positive integer")
		return
	}
	t, err := h.uc.GetTestByID(r.Context(), id)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, t)
}

type createAdminUserRequest struct {
	TelegramID         int64  `json:"telegram_id"`
	Role               string `json:"role"`
	SubscriptionStatus string `json:"subscription_status"`
}

// CreateAdminUser handles POST /api/v1/admin/users.
func (h *AdminHandler) CreateAdminUser(w http.ResponseWriter, r *http.Request) {
	var req createAdminUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.TelegramID == 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "telegram_id is required")
		return
	}
	role := user.Role(req.Role)
	if role == "" {
		role = user.RoleStudent
	}
	subStatus := user.SubscriptionStatus(req.SubscriptionStatus)
	if subStatus == "" {
		subStatus = user.SubscriptionStatusInactive
	}
	u, err := h.uc.CreateUser(r.Context(), req.TelegramID, role, subStatus)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, u)
}

// DeleteAdminUser handles DELETE /api/v1/admin/users/{telegram_id}.
func (h *AdminHandler) DeleteAdminUser(w http.ResponseWriter, r *http.Request) {
	telegramID, err := strconv.ParseInt(r.PathValue("telegram_id"), 10, 64)
	if err != nil || telegramID <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "telegram_id must be a valid positive integer")
		return
	}
	if err := h.uc.DeleteUser(r.Context(), telegramID); err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]int64{"telegram_id": telegramID})
}

// GetTeacherStudents handles GET /api/v1/admin/users/{telegram_id}/students.
func (h *AdminHandler) GetTeacherStudents(w http.ResponseWriter, r *http.Request) {
	telegramID, err := strconv.ParseInt(r.PathValue("telegram_id"), 10, 64)
	if err != nil || telegramID <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "telegram_id must be a valid positive integer")
		return
	}
	students, err := h.uc.GetStudentsByTeacher(r.Context(), telegramID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	if students == nil {
		students = []*user.User{}
	}
	respond.JSON(w, http.StatusOK, map[string]interface{}{"data": students, "total": len(students)})
}

// GetAdminUser handles GET /api/v1/admin/users/{telegram_id}.
func (h *AdminHandler) GetAdminUser(w http.ResponseWriter, r *http.Request) {
	telegramID, err := strconv.ParseInt(r.PathValue("telegram_id"), 10, 64)
	if err != nil || telegramID <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "telegram_id must be a valid positive integer")
		return
	}
	u, err := h.uc.GetUser(r.Context(), telegramID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, u)
}

type updateAdminUserRequest struct {
	Role               *string `json:"role"`
	SubscriptionStatus *string `json:"subscription_status"`
}

// UpdateAdminUser handles PUT /api/v1/admin/users/{telegram_id}.
func (h *AdminHandler) UpdateAdminUser(w http.ResponseWriter, r *http.Request) {
	telegramID, err := strconv.ParseInt(r.PathValue("telegram_id"), 10, 64)
	if err != nil || telegramID <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "telegram_id must be a valid positive integer")
		return
	}
	var req updateAdminUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	var rolePtr *user.Role
	if req.Role != nil {
		r := user.Role(*req.Role)
		rolePtr = &r
	}
	var subStatusPtr *user.SubscriptionStatus
	if req.SubscriptionStatus != nil {
		s := user.SubscriptionStatus(*req.SubscriptionStatus)
		subStatusPtr = &s
	}
	u, err := h.uc.UpdateUserState(r.Context(), telegramID, rolePtr, subStatusPtr)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, u)
}

// GetFeedbackList handles GET /api/v1/admin/feedback.
func (h *AdminHandler) GetFeedbackList(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	items, total, err := h.uc.GetFeedback(r.Context(), limit, offset)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	if items == nil {
		items = []*feedbackDomain.Feedback{}
	}
	respond.JSON(w, http.StatusOK, map[string]any{"data": items, "total": total})
}

// GetFeedbackItem handles GET /api/v1/admin/feedback/{id}.
func (h *AdminHandler) GetFeedbackItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	item, err := h.uc.GetFeedbackByID(r.Context(), id)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, item)
}

// GetInviteLinkList handles GET /api/v1/admin/invite-links.
func (h *AdminHandler) GetInviteLinkList(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	items, total, err := h.uc.GetInviteLinks(r.Context(), limit, offset)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"data": items, "total": total})
}

// GetInviteLinkItem handles GET /api/v1/admin/invite-links/{id}.
func (h *AdminHandler) GetInviteLinkItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	item, err := h.uc.GetInviteLinkByID(r.Context(), id)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, item)
}

// GetCorporatePurchases handles GET /api/v1/admin/corporate-purchases.
func (h *AdminHandler) GetCorporatePurchases(w http.ResponseWriter, r *http.Request) {
	managerIDStr := r.URL.Query().Get("manager_id")
	var purchases interface{}
	var err error
	if managerIDStr != "" {
		managerID, parseErr := strconv.ParseInt(managerIDStr, 10, 64)
		if parseErr != nil || managerID <= 0 {
			respond.Error(w, http.StatusBadRequest, "bad_request", "manager_id must be a valid positive integer")
			return
		}
		purchases, err = h.managerUC.GetPurchases(r.Context(), managerID)
	} else {
		purchases, err = h.managerUC.GetAllPurchases(r.Context())
	}
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]interface{}{"data": purchases})
}

// GetCorporatePurchase handles GET /api/v1/admin/corporate-purchases/{payment_id}.
func (h *AdminHandler) GetCorporatePurchase(w http.ResponseWriter, r *http.Request) {
	paymentID := r.PathValue("payment_id")
	if paymentID == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "payment_id is required")
		return
	}
	purchase, err := h.managerUC.GetPurchaseByID(r.Context(), paymentID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, purchase)
}

// GetCorporatePurchasePDF handles GET /api/v1/admin/corporate-purchases/{payment_id}/pdf.
func (h *AdminHandler) GetCorporatePurchasePDF(w http.ResponseWriter, r *http.Request) {
	paymentID := r.PathValue("payment_id")
	if paymentID == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "payment_id is required")
		return
	}
	// Admin can regenerate PDF for any purchase (no owner check).
	data, err := h.managerUC.GeneratePurchasePDF(r.Context(), 0, paymentID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="corporate_%s.pdf"`, paymentID))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetCorporateGroups handles GET /api/v1/admin/corporate-groups.
func (h *AdminHandler) GetCorporateGroups(w http.ResponseWriter, r *http.Request) {
	purchaseIDFilter := r.URL.Query().Get("purchase_id")
	managerIDStr := r.URL.Query().Get("manager_id")

	if purchaseIDFilter != "" {
		groups, err := h.managerUC.GetGroupsByPurchase(r.Context(), purchaseIDFilter)
		if err != nil {
			respond.ErrorFrom(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, map[string]interface{}{"data": groups})
		return
	}
	if managerIDStr != "" {
		managerID, parseErr := strconv.ParseInt(managerIDStr, 10, 64)
		if parseErr != nil || managerID <= 0 {
			respond.Error(w, http.StatusBadRequest, "bad_request", "manager_id must be a valid positive integer")
			return
		}
		groups, err := h.managerUC.GetGroupsByManager(r.Context(), managerID)
		if err != nil {
			respond.ErrorFrom(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, map[string]interface{}{"data": groups})
		return
	}
	groups, err := h.managerUC.GetAllGroups(r.Context())
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]interface{}{"data": groups})
}

// GetCorporateGroup handles GET /api/v1/admin/corporate-groups/{id}.
func (h *AdminHandler) GetCorporateGroup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "id is required")
		return
	}
	group, err := h.managerUC.GetGroupByID(r.Context(), id)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, group)
}
