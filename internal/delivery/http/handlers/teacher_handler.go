package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/vladkonst/mnemonics/internal/delivery/http/respond"
	teacherUC "github.com/vladkonst/mnemonics/internal/usecase/teacher"
)

// TeacherHandler handles teacher-specific HTTP endpoints.
type TeacherHandler struct {
	uc *teacherUC.UseCase
}

// NewTeacherHandler creates a new TeacherHandler.
func NewTeacherHandler(uc *teacherUC.UseCase) *TeacherHandler {
	return &TeacherHandler{uc: uc}
}

// GetStudents handles GET /api/v1/teachers/{teacher_id}/students.
func (h *TeacherHandler) GetStudents(w http.ResponseWriter, r *http.Request) {
	teacherID, err := parseTeacherID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	result, err := h.uc.GetStudents(r.Context(), teacherID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, result)
}

// GetStudentProgress handles GET /api/v1/teachers/{teacher_id}/students/{student_id}/progress.
func (h *TeacherHandler) GetStudentProgress(w http.ResponseWriter, r *http.Request) {
	teacherID, err := parseTeacherID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	studentID, err := parseStudentID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	result, err := h.uc.GetStudentProgress(r.Context(), teacherID, studentID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, result)
}

// GetStatistics handles GET /api/v1/teachers/{teacher_id}/statistics.
func (h *TeacherHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	teacherID, err := parseTeacherID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	result, err := h.uc.GetStatistics(r.Context(), teacherID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, result)
}

// ClaimCorporateGroup handles POST /api/v1/teachers/{teacher_id}/groups/claim.
func (h *TeacherHandler) ClaimCorporateGroup(w http.ResponseWriter, r *http.Request) {
	teacherID, err := parseTeacherID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	var body struct {
		JoinCode string `json:"join_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	if body.JoinCode == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "join_code is required")
		return
	}

	group, err := h.uc.ClaimCorporateGroup(r.Context(), teacherID, body.JoinCode)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, group)
}

// GetCorporateGroups handles GET /api/v1/teachers/{teacher_id}/groups.
func (h *TeacherHandler) GetCorporateGroups(w http.ResponseWriter, r *http.Request) {
	teacherID, err := parseTeacherID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	groups, err := h.uc.GetCorporateGroupsWithStats(r.Context(), teacherID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"groups": groups,
		"total":  len(groups),
	})
}

// GetCorporateGroupStudents handles GET /api/v1/teachers/{teacher_id}/groups/{group_id}/students.
func (h *TeacherHandler) GetCorporateGroupStudents(w http.ResponseWriter, r *http.Request) {
	teacherID, err := parseTeacherID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	groupID := r.PathValue("group_id")
	if groupID == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "group_id is required")
		return
	}

	result, err := h.uc.GetCorporateGroupStudents(r.Context(), teacherID, groupID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, result)
}

// GetCorporateGroupStats handles GET /api/v1/teachers/{teacher_id}/groups/{group_id}/statistics.
func (h *TeacherHandler) GetCorporateGroupStats(w http.ResponseWriter, r *http.Request) {
	teacherID, err := parseTeacherID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	groupID := r.PathValue("group_id")
	if groupID == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "group_id is required")
		return
	}

	result, err := h.uc.GetCorporateGroupStats(r.Context(), teacherID, groupID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, result)
}

// RenameCorporateGroup handles PATCH /api/v1/teachers/{teacher_id}/groups/{group_id}.
func (h *TeacherHandler) RenameCorporateGroup(w http.ResponseWriter, r *http.Request) {
	teacherID, err := parseTeacherID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	groupID := r.PathValue("group_id")
	if groupID == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "group_id is required")
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	if err := h.uc.RenameCorporateGroup(r.Context(), teacherID, groupID, body.Name); err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// parseStudentID extracts and validates the student_id path parameter.
func parseStudentID(r *http.Request) (int64, error) {
	raw := r.PathValue("student_id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("student_id must be a valid positive integer")
	}
	return id, nil
}
