package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/vladkonst/mnemonics/internal/delivery/http/respond"
	contentUC "github.com/vladkonst/mnemonics/internal/usecase/content"
	"github.com/vladkonst/mnemonics/internal/domain/progress"
)

// ContentHandler handles content-related HTTP endpoints.
type ContentHandler struct {
	uc *contentUC.UseCase
}

// NewContentHandler creates a new ContentHandler.
func NewContentHandler(uc *contentUC.UseCase) *ContentHandler {
	return &ContentHandler{uc: uc}
}

// GetModules handles GET /api/v1/content/modules?user_id=...
func (h *ContentHandler) GetModules(w http.ResponseWriter, r *http.Request) {
	userID, err := parseQueryUserID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	modules, err := h.uc.GetModules(r.Context(), userID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"modules": modules,
	})
}

// GetModuleThemes handles GET /api/v1/content/modules/{module_id}/themes?user_id=...
func (h *ContentHandler) GetModuleThemes(w http.ResponseWriter, r *http.Request) {
	moduleID, err := parseModuleID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	userID, err := parseQueryUserID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	result, err := h.uc.GetModuleThemes(r.Context(), moduleID, userID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, result)
}

// createStudySessionRequest is the JSON body for POST /api/v1/users/{user_id}/study-sessions.
type createStudySessionRequest struct {
	ThemeID int `json:"theme_id"`
}

// CreateStudySession handles POST /api/v1/users/{user_id}/study-sessions.
func (h *ContentHandler) CreateStudySession(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	var req createStudySessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.ThemeID <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "theme_id is required")
		return
	}

	result, err := h.uc.CreateStudySession(r.Context(), userID, req.ThemeID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/users/%d/study-sessions/%s", userID, result.SessionID))
	respond.JSON(w, http.StatusCreated, result)
}

// startTestAttemptRequest is the JSON body for POST /api/v1/users/{user_id}/test-attempts.
type startTestAttemptRequest struct {
	ThemeID int `json:"theme_id"`
}

// StartTestAttempt handles POST /api/v1/users/{user_id}/test-attempts.
func (h *ContentHandler) StartTestAttempt(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	var req startTestAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.ThemeID <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "theme_id is required")
		return
	}

	attempt, err := h.uc.StartTestAttempt(r.Context(), userID, req.ThemeID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/users/%d/test-attempts/%s", userID, attempt.AttemptID))
	respond.JSON(w, http.StatusCreated, attempt)
}

// submitTestAttemptRequest is the JSON body for PUT /api/v1/users/{user_id}/test-attempts/{attempt_id}.
type submitTestAttemptRequest struct {
	Answers []answerItemRequest `json:"answers"`
}

type answerItemRequest struct {
	QuestionID int    `json:"question_id"`
	Answer     string `json:"answer"`
}

// SubmitTestAttempt handles PUT /api/v1/users/{user_id}/test-attempts/{attempt_id}.
func (h *ContentHandler) SubmitTestAttempt(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	attemptID := r.PathValue("attempt_id")
	if attemptID == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "attempt_id is required")
		return
	}

	var req submitTestAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}

	answers := make([]progress.AnswerItem, 0, len(req.Answers))
	for _, a := range req.Answers {
		answers = append(answers, progress.AnswerItem{
			QuestionID: a.QuestionID,
			Answer:     a.Answer,
		})
	}

	result, err := h.uc.SubmitTestAttempt(r.Context(), userID, attemptID, answers)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, result)
}

// CheckThemeAccess handles GET /api/v1/users/{user_id}/theme/{theme_id}/access.
func (h *ContentHandler) CheckThemeAccess(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	themeID, err := parseThemeIDPath(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	result, err := h.uc.CheckThemeAccess(r.Context(), userID, themeID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, result)
}

// parseQueryUserID parses user_id from the query string.
func parseQueryUserID(r *http.Request) (int64, error) {
	raw := r.URL.Query().Get("user_id")
	if raw == "" {
		return 0, fmt.Errorf("user_id query parameter is required")
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("user_id must be a valid positive integer")
	}
	return id, nil
}

// parseModuleID extracts and validates the module_id path parameter.
func parseModuleID(r *http.Request) (int, error) {
	raw := r.PathValue("module_id")
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("module_id must be a valid positive integer")
	}
	return id, nil
}

// parseThemeIDPath extracts and validates the theme_id path parameter.
func parseThemeIDPath(r *http.Request) (int, error) {
	raw := r.PathValue("theme_id")
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("theme_id must be a valid positive integer")
	}
	return id, nil
}
