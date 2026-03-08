// Package progress provides use cases for user progress tracking and analytics.
package progress

import (
	"context"

	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
	"github.com/vladkonst/mnemonics/internal/domain/progress"
)

// ── Result types ─────────────────────────────────────────────────────────────

// ThemeProgressItem summarises progress for a single theme.
type ThemeProgressItem struct {
	ThemeID       int
	ThemeName     string
	Status        progress.Status
	Score         *int
	AttemptCount  int
	CompletedAt   interface{} // *time.Time or nil
}

// ModuleSummary summarises progress across all themes in a module.
type ModuleSummary struct {
	ModuleID        int
	ModuleName      string
	TotalThemes     int
	CompletedThemes int
	AverageScore    *int
}

// UserProgressResult contains overall user progress statistics.
type UserProgressResult struct {
	UserID          int64
	TotalThemes     int
	CompletedThemes int
	AverageScore    *int
	RecentActivity  []*progress.UserProgress
	ModuleSummaries []*ModuleSummary
}

// ModuleProgressResult contains per-module progress details.
type ModuleProgressResult struct {
	ModuleID    int
	ModuleName  string
	Themes      []*ThemeProgressItem
	Completed   int
	Total       int
	AverageScore *int
}

// ── UseCase ──────────────────────────────────────────────────────────────────

// UseCase orchestrates progress analytics.
type UseCase struct {
	progress interfaces.ProgressRepository
	attempts interfaces.TestAttemptRepository
	tests    interfaces.TestRepository
	themes   interfaces.ThemeRepository
	modules  interfaces.ModuleRepository
}

// NewUseCase creates a new progress UseCase.
func NewUseCase(
	progress interfaces.ProgressRepository,
	attempts interfaces.TestAttemptRepository,
	tests interfaces.TestRepository,
	themes interfaces.ThemeRepository,
	modules interfaces.ModuleRepository,
) *UseCase {
	return &UseCase{
		progress: progress,
		attempts: attempts,
		tests:    tests,
		themes:   themes,
		modules:  modules,
	}
}

// GetUserProgress returns overall progress stats for a user.
func (uc *UseCase) GetUserProgress(ctx context.Context, userID int64) (*UserProgressResult, error) {
	allProgress, err := uc.progress.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	completed := 0
	scoreSum := 0
	scoreCount := 0
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

	// Build module summaries.
	modules, err := uc.modules.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	summaries := make([]*ModuleSummary, 0, len(modules))
	for _, m := range modules {
		themes, err := uc.themes.GetByModuleID(ctx, m.ID)
		if err != nil {
			return nil, err
		}

		modCompleted := 0
		modScoreSum := 0
		modScoreCount := 0
		for _, t := range themes {
			p, err := uc.progress.GetByUserAndTheme(ctx, userID, t.ID)
			if err != nil {
				continue
			}
			if p.IsCompleted() {
				modCompleted++
				if p.Score != nil {
					modScoreSum += *p.Score
					modScoreCount++
				}
			}
		}

		var modAvg *int
		if modScoreCount > 0 {
			avg := modScoreSum / modScoreCount
			modAvg = &avg
		}

		summaries = append(summaries, &ModuleSummary{
			ModuleID:        m.ID,
			ModuleName:      m.Name,
			TotalThemes:     len(themes),
			CompletedThemes: modCompleted,
			AverageScore:    modAvg,
		})
	}

	// Recent activity: last 10 progress entries sorted by UpdatedAt descending.
	recent := allProgress
	if len(recent) > 10 {
		recent = recent[len(recent)-10:]
	}

	return &UserProgressResult{
		UserID:          userID,
		TotalThemes:     len(allProgress),
		CompletedThemes: completed,
		AverageScore:    avgScore,
		RecentActivity:  recent,
		ModuleSummaries: summaries,
	}, nil
}

// GetModuleProgress returns detailed per-theme progress for a user in a module.
func (uc *UseCase) GetModuleProgress(ctx context.Context, userID int64, moduleID int) (*ModuleProgressResult, error) {
	mod, err := uc.modules.GetByID(ctx, moduleID)
	if err != nil {
		return nil, err
	}

	themes, err := uc.themes.GetByModuleID(ctx, moduleID)
	if err != nil {
		return nil, err
	}

	items := make([]*ThemeProgressItem, 0, len(themes))
	completed := 0
	scoreSum := 0
	scoreCount := 0

	for _, t := range themes {
		item := &ThemeProgressItem{
			ThemeID:   t.ID,
			ThemeName: t.Name,
			Status:    progress.StatusStarted,
		}

		p, err := uc.progress.GetByUserAndTheme(ctx, userID, t.ID)
		if err == nil && p != nil {
			item.Status = p.Status
			item.Score = p.Score
			item.CompletedAt = p.CompletedAt

			if p.IsCompleted() {
				completed++
				if p.Score != nil {
					scoreSum += *p.Score
					scoreCount++
				}
			}

			// Count attempts.
			attempts, err := uc.attempts.GetByUserAndTheme(ctx, userID, t.ID)
			if err == nil {
				item.AttemptCount = len(attempts)
			}
		}

		items = append(items, item)
	}

	var avgScore *int
	if scoreCount > 0 {
		avg := scoreSum / scoreCount
		avgScore = &avg
	}

	return &ModuleProgressResult{
		ModuleID:     mod.ID,
		ModuleName:   mod.Name,
		Themes:       items,
		Completed:    completed,
		Total:        len(themes),
		AverageScore: avgScore,
	}, nil
}
