package handlers

import (
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

// parseStudentID extracts and validates the student_id path parameter.
func parseStudentID(r *http.Request) (int64, error) {
	raw := r.PathValue("student_id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("student_id must be a valid positive integer")
	}
	return id, nil
}
