// Package teacher provides use cases for teacher-specific operations.
package teacher

import (
	"context"

	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
	"github.com/vladkonst/mnemonics/internal/domain/progress"
	"github.com/vladkonst/mnemonics/internal/domain/user"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// ── Result types ─────────────────────────────────────────────────────────────

// StudentSummary is a brief summary of a student's state.
type StudentSummary struct {
	*user.User
	CompletedThemes int  `json:"completed_themes"`
	AverageScore    *int `json:"average_score,omitempty"`
}

// StudentsResult contains the list of students for a teacher.
type StudentsResult struct {
	TeacherID int64             `json:"teacher_id"`
	Students  []*StudentSummary `json:"students"`
	Total     int               `json:"total"`
}

// ThemeProgressItem is a progress record for a single theme.
type ThemeProgressItem struct {
	ThemeID      int             `json:"theme_id"`
	ThemeName    string          `json:"theme_name"`
	Status       progress.Status `json:"status"`
	Score        *int            `json:"score,omitempty"`
	AttemptCount int             `json:"attempt_count"`
}

// StudentProgressResult contains a detailed progress view for a specific student.
type StudentProgressResult struct {
	Student         *user.User           `json:"student"`
	CompletedThemes int                  `json:"completed_themes"`
	TotalThemes     int                  `json:"total_themes"`
	AverageScore    *int                 `json:"average_score,omitempty"`
	ThemeProgress   []*ThemeProgressItem `json:"theme_progress"`
}

// GroupStatItem is a per-student summary row for teacher statistics.
type GroupStatItem struct {
	Student         *user.User `json:"student"`
	CompletedThemes int        `json:"completed_themes"`
	AverageScore    *int       `json:"average_score,omitempty"`
}

// GroupStatisticsResult contains group-level analytics for a teacher's students.
type GroupStatisticsResult struct {
	TeacherID      int64            `json:"teacher_id"`
	TotalStudents  int              `json:"total_students"`
	AverageScore   *int             `json:"average_score,omitempty"`
	CompletionRate float64          `json:"completion_rate"`
	StudentStats   []*GroupStatItem `json:"student_stats"`
}

// ── UseCase ──────────────────────────────────────────────────────────────────

// UseCase orchestrates teacher operations.
type UseCase struct {
	teacherStudents interfaces.TeacherStudentRepository
	progress        interfaces.ProgressRepository
	attempts        interfaces.TestAttemptRepository
	modules         interfaces.ModuleRepository
	themes          interfaces.ThemeRepository
	users           interfaces.UserRepository
}

// NewUseCase creates a new teacher UseCase.
func NewUseCase(
	teacherStudents interfaces.TeacherStudentRepository,
	progress interfaces.ProgressRepository,
	attempts interfaces.TestAttemptRepository,
	modules interfaces.ModuleRepository,
	themes interfaces.ThemeRepository,
	users interfaces.UserRepository,
) *UseCase {
	return &UseCase{
		teacherStudents: teacherStudents,
		progress:        progress,
		attempts:        attempts,
		modules:         modules,
		themes:          themes,
		users:           users,
	}
}

// GetStudents returns all students for a teacher with completion summaries.
func (uc *UseCase) GetStudents(ctx context.Context, teacherID int64) (*StudentsResult, error) {
	teacher, err := uc.users.GetByID(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	if !teacher.IsTeacher() {
		return nil, apperrors.ErrNotTeacher
	}

	students, err := uc.teacherStudents.GetStudentsByTeacher(ctx, teacherID)
	if err != nil {
		return nil, err
	}

	summaries := make([]*StudentSummary, 0, len(students))
	for _, s := range students {
		allProgress, err := uc.progress.GetByUser(ctx, s.TelegramID)
		if err != nil {
			return nil, err
		}

		completed := 0
		scoreSum, scoreCount := 0, 0
		for _, p := range allProgress {
			if p.IsCompleted() {
				completed++
				if p.Score != nil {
					scoreSum += *p.Score
					scoreCount++
				}
			}
		}

		var avgScore *int
		if scoreCount > 0 {
			avg := scoreSum / scoreCount
			avgScore = &avg
		}

		summaries = append(summaries, &StudentSummary{
			User:            s,
			CompletedThemes: completed,
			AverageScore:    avgScore,
		})
	}

	return &StudentsResult{
		TeacherID: teacherID,
		Students:  summaries,
		Total:     len(summaries),
	}, nil
}

// GetStudentProgress returns detailed progress for a specific student,
// verifying the student belongs to the teacher first.
func (uc *UseCase) GetStudentProgress(ctx context.Context, teacherID, studentID int64) (*StudentProgressResult, error) {
	teacher, err := uc.users.GetByID(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	if !teacher.IsTeacher() {
		return nil, apperrors.ErrNotTeacher
	}

	// Verify student belongs to this teacher.
	belongs, err := uc.teacherStudents.IsTeacherStudent(ctx, teacherID, studentID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperrors.ErrNotYourStudent
	}

	student, err := uc.users.GetByID(ctx, studentID)
	if err != nil {
		return nil, err
	}

	// Gather all themes across modules.
	modules, err := uc.modules.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var items []*ThemeProgressItem
	totalThemes := 0
	completed := 0
	scoreSum, scoreCount := 0, 0

	for _, m := range modules {
		themes, err := uc.themes.GetByModuleID(ctx, m.ID)
		if err != nil {
			return nil, err
		}
		totalThemes += len(themes)

		for _, t := range themes {
			item := &ThemeProgressItem{
				ThemeID:   t.ID,
				ThemeName: t.Name,
				Status:    progress.StatusStarted,
			}

			p, err := uc.progress.GetByUserAndTheme(ctx, studentID, t.ID)
			if err == nil && p != nil {
				item.Status = p.Status
				item.Score = p.Score
				if p.IsCompleted() {
					completed++
					if p.Score != nil {
						scoreSum += *p.Score
						scoreCount++
					}
				}
			}

			attempts, err := uc.attempts.GetByUserAndTheme(ctx, studentID, t.ID)
			if err == nil {
				item.AttemptCount = len(attempts)
			}

			items = append(items, item)
		}
	}

	var avgScore *int
	if scoreCount > 0 {
		avg := scoreSum / scoreCount
		avgScore = &avg
	}

	return &StudentProgressResult{
		Student:         student,
		CompletedThemes: completed,
		TotalThemes:     totalThemes,
		AverageScore:    avgScore,
		ThemeProgress:   items,
	}, nil
}

// GetStatistics returns group-level statistics for all of a teacher's students.
func (uc *UseCase) GetStatistics(ctx context.Context, teacherID int64) (*GroupStatisticsResult, error) {
	teacher, err := uc.users.GetByID(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	if !teacher.IsTeacher() {
		return nil, apperrors.ErrNotTeacher
	}

	students, err := uc.teacherStudents.GetStudentsByTeacher(ctx, teacherID)
	if err != nil {
		return nil, err
	}

	// Count total themes once.
	modules, err := uc.modules.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	totalThemesGlobal := 0
	for _, m := range modules {
		themes, err := uc.themes.GetByModuleID(ctx, m.ID)
		if err != nil {
			return nil, err
		}
		totalThemesGlobal += len(themes)
	}

	statItems := make([]*GroupStatItem, 0, len(students))
	groupScoreSum, groupScoreCount := 0, 0
	totalCompleted := 0

	for _, s := range students {
		allProgress, err := uc.progress.GetByUser(ctx, s.TelegramID)
		if err != nil {
			return nil, err
		}

		completed := 0
		scoreSum, scoreCount := 0, 0
		for _, p := range allProgress {
			if p.IsCompleted() {
				completed++
				if p.Score != nil {
					scoreSum += *p.Score
					scoreCount++
					groupScoreSum += *p.Score
					groupScoreCount++
				}
			}
		}
		totalCompleted += completed

		var avgScore *int
		if scoreCount > 0 {
			avg := scoreSum / scoreCount
			avgScore = &avg
		}

		statItems = append(statItems, &GroupStatItem{
			Student:         s,
			CompletedThemes: completed,
			AverageScore:    avgScore,
		})
	}

	var groupAvg *int
	if groupScoreCount > 0 {
		avg := groupScoreSum / groupScoreCount
		groupAvg = &avg
	}

	var completionRate float64
	if len(students) > 0 && totalThemesGlobal > 0 {
		completionRate = float64(totalCompleted) / float64(len(students)*totalThemesGlobal) * 100
	}

	return &GroupStatisticsResult{
		TeacherID:      teacherID,
		TotalStudents:  len(students),
		AverageScore:   groupAvg,
		CompletionRate: completionRate,
		StudentStats:   statItems,
	}, nil
}
