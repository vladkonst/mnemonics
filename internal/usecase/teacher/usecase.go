// Package teacher provides use cases for teacher-specific operations.
package teacher

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
	"github.com/vladkonst/mnemonics/internal/domain/progress"
	"github.com/vladkonst/mnemonics/internal/domain/subscription"
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
	corporateGroups interfaces.CorporateGroupRepository
	inviteLinks     interfaces.InviteLinkRepository
	subscriptions   interfaces.SubscriptionRepository
}

// NewUseCase creates a new teacher UseCase.
func NewUseCase(
	teacherStudents interfaces.TeacherStudentRepository,
	progress interfaces.ProgressRepository,
	attempts interfaces.TestAttemptRepository,
	modules interfaces.ModuleRepository,
	themes interfaces.ThemeRepository,
	users interfaces.UserRepository,
	corporateGroups interfaces.CorporateGroupRepository,
	inviteLinks interfaces.InviteLinkRepository,
	subscriptions interfaces.SubscriptionRepository,
) *UseCase {
	return &UseCase{
		teacherStudents: teacherStudents,
		progress:        progress,
		attempts:        attempts,
		modules:         modules,
		themes:          themes,
		users:           users,
		corporateGroups: corporateGroups,
		inviteLinks:     inviteLinks,
		subscriptions:   subscriptions,
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

// ── Corporate group methods ───────────────────────────────────────────────────

// ClaimCorporateGroup lets a teacher claim an unclaimed corporate group via a join code.
// It also grants the teacher an active subscription for the group's semester duration.
func (uc *UseCase) ClaimCorporateGroup(ctx context.Context, teacherID int64, joinCode string) (*subscription.CorporateGroup, error) {
	group, err := uc.corporateGroups.GetByJoinCode(ctx, joinCode)
	if err != nil {
		return nil, err
	}
	if err := uc.corporateGroups.ClaimByTeacher(ctx, group.ID, teacherID); err != nil {
		return nil, err
	}

	// Set teacher role.
	u, err := uc.users.GetByID(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	u.SetRole(user.RoleTeacher)
	if err := uc.users.Update(ctx, u); err != nil {
		return nil, err
	}

	// Grant subscription if not already active.
	existing, err := uc.subscriptions.GetActiveByUserID(ctx, teacherID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}
	if existing == nil || !existing.IsActive() {
		expiresAt := time.Now().UTC().AddDate(0, group.Semesters*5, 0)
		paymentID := fmt.Sprintf("corp-teacher-%s-%d-%d", group.ID, teacherID, time.Now().UnixNano())
		sub := &subscription.Subscription{
			PaymentID: paymentID,
			UserID:    teacherID,
			Type:      subscription.SubscriptionTypeUniversity,
			Status:    subscription.SubscriptionPlanStatusActive,
			ExpiresAt: &expiresAt,
			CreatedAt: time.Now().UTC(),
		}
		if err := uc.subscriptions.Create(ctx, sub); err != nil {
			return nil, err
		}
	}

	// Return updated group.
	return uc.corporateGroups.GetByID(ctx, group.ID)
}

// GetCorporateGroupsWithStats returns all corporate groups claimed by a teacher, with student counts.
func (uc *UseCase) GetCorporateGroupsWithStats(ctx context.Context, teacherID int64) ([]*subscription.CorporateGroupWithStats, error) {
	groups, err := uc.corporateGroups.GetByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	result := make([]*subscription.CorporateGroupWithStats, len(groups))
	for i, g := range groups {
		count, _ := uc.inviteLinks.CountActivations(ctx, g.StudentLinkID)
		result[i] = &subscription.CorporateGroupWithStats{CorporateGroup: g, StudentCount: count}
	}
	return result, nil
}

// GetCorporateGroupStudents returns the students of a specific corporate group,
// verifying the requesting teacher actually owns that group.
func (uc *UseCase) GetCorporateGroupStudents(ctx context.Context, teacherID int64, groupID string) (*StudentsResult, error) {
	group, err := uc.corporateGroups.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.TeacherID == nil || *group.TeacherID != teacherID {
		return nil, apperrors.ErrForbidden
	}

	activations, err := uc.inviteLinks.GetActivations(ctx, group.StudentLinkID)
	if err != nil {
		return nil, err
	}

	summaries := make([]*StudentSummary, 0, len(activations))
	for _, act := range activations {
		s, err := uc.users.GetByID(ctx, act.UserID)
		if err != nil {
			continue
		}
		allProgress, err := uc.progress.GetByUser(ctx, s.TelegramID)
		if err != nil {
			continue
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

// GetCorporateGroupStats returns aggregate statistics for a corporate group.
func (uc *UseCase) GetCorporateGroupStats(ctx context.Context, teacherID int64, groupID string) (*GroupStatisticsResult, error) {
	studentsResult, err := uc.GetCorporateGroupStudents(ctx, teacherID, groupID)
	if err != nil {
		return nil, err
	}

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

	statItems := make([]*GroupStatItem, 0, len(studentsResult.Students))
	groupScoreSum, groupScoreCount := 0, 0
	totalCompleted := 0

	for _, s := range studentsResult.Students {
		totalCompleted += s.CompletedThemes
		if s.AverageScore != nil {
			groupScoreSum += *s.AverageScore
			groupScoreCount++
		}
		statItems = append(statItems, &GroupStatItem{
			Student:         s.User,
			CompletedThemes: s.CompletedThemes,
			AverageScore:    s.AverageScore,
		})
	}

	var groupAvg *int
	if groupScoreCount > 0 {
		avg := groupScoreSum / groupScoreCount
		groupAvg = &avg
	}

	var completionRate float64
	n := len(studentsResult.Students)
	if n > 0 && totalThemesGlobal > 0 {
		completionRate = float64(totalCompleted) / float64(n*totalThemesGlobal) * 100
	}

	return &GroupStatisticsResult{
		TeacherID:      teacherID,
		TotalStudents:  n,
		AverageScore:   groupAvg,
		CompletionRate: completionRate,
		StudentStats:   statItems,
	}, nil
}

// RenameCorporateGroup updates the display name of a corporate group the teacher owns.
func (uc *UseCase) RenameCorporateGroup(ctx context.Context, teacherID int64, groupID, newName string) error {
	group, err := uc.corporateGroups.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.TeacherID == nil || *group.TeacherID != teacherID {
		return apperrors.ErrForbidden
	}
	newName = strings.TrimSpace(newName)
	if newName == "" || len([]rune(newName)) > 100 {
		return apperrors.ErrInvalidInput
	}
	return uc.corporateGroups.UpdateName(ctx, groupID, newName)
}
