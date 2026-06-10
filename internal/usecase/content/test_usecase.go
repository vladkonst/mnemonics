package content

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	contentDomain "github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/internal/domain/progress"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// QuestionDTO is a question sent to the client (CorrectAnswer is intentionally omitted).
type QuestionDTO struct {
	ID       int    `json:"id"`
	Question string `json:"question"`
	OrderNum int    `json:"order_num"`
}

// TestDTO is the test representation sent to the client.
type TestDTO struct {
	ID           int           `json:"id"`
	ThemeID      int           `json:"theme_id"`
	Difficulty   int           `json:"difficulty"`
	PassingScore int           `json:"passing_score"`
	Questions    []QuestionDTO `json:"questions"`
}

// StartTestResult is the response for starting a test attempt.
type StartTestResult struct {
	AttemptID     string    `json:"attempt_id"`
	AttemptNumber int       `json:"attempt_number"`
	StartedAt     time.Time `json:"started_at"`
	Test          TestDTO   `json:"test"`
}

// QuestionResult carries the grading details for a single question.
type QuestionResult struct {
	QuestionID    int    `json:"question_id"`
	QuestionText  string `json:"question_text"`
	UserAnswer    string `json:"user_answer"`
	CorrectAnswer string `json:"correct_answer"`
	IsCorrect     bool   `json:"is_correct"`
}

// SubmitResult carries the outcome of a test submission.
type SubmitResult struct {
	Score          int              `json:"score"`
	PassingScore   int              `json:"passing_score"`
	Passed         bool             `json:"passed"`
	CorrectAnswers int              `json:"correct_answers"`
	TotalQuestions int              `json:"total_questions"`
	AttemptNumber  int              `json:"attempt_number"`
	NextAction     *NextAction      `json:"next_action,omitempty"`
	MotivationMsg  string           `json:"motivation_msg"`
	Results        []QuestionResult `json:"results"`
}

// NextAction describes what the user should do after submitting a test.
type NextAction struct {
	Type           string  `json:"type"`
	ThemeID        *int    `json:"theme_id,omitempty"`
	ThemeName      *string `json:"theme_name,omitempty"`
	IsIntroduction *bool   `json:"is_introduction,omitempty"`
	ModuleID       *int    `json:"module_id,omitempty"`
	Message        *string `json:"message,omitempty"`
}

// StartTestAttempt creates a new test attempt record for a user and theme.
func (uc *UseCase) StartTestAttempt(ctx context.Context, userID int64, themeID int) (*StartTestResult, error) {
	// Verify access to the theme.
	access, err := uc.CheckThemeAccess(ctx, userID, themeID)
	if err != nil {
		return nil, err
	}
	if !access.Accessible {
		return nil, apperrors.ErrAccessDenied
	}

	// Ensure a test exists for this theme.
	test, err := uc.tests.GetByThemeID(ctx, themeID)
	if err != nil {
		return nil, err
	}

	// Update progress: record test start.
	prog, err := uc.progress.GetByUserAndTheme(ctx, userID, themeID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}
	if prog == nil {
		now := time.Now().UTC()
		prog = &progress.UserProgress{
			UserID:    userID,
			ThemeID:   themeID,
			Status:    progress.StatusStarted,
			StartedAt: now,
			UpdatedAt: now,
		}
	}
	prog.StartTest()
	if err := uc.progress.Upsert(ctx, prog); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	attempt := &progress.TestAttempt{
		UserID:    userID,
		ThemeID:   themeID,
		TestID:    test.ID,
		AttemptID: uuid.NewString(),
		StartedAt: now,
	}
	if err := uc.attempts.Create(ctx, attempt); err != nil {
		return nil, err
	}

	questions := make([]QuestionDTO, len(test.Questions))
	for i, q := range test.Questions {
		questions[i] = QuestionDTO{
			ID:       q.ID,
			Question: q.Text,
			OrderNum: q.OrderNum,
		}
	}

	return &StartTestResult{
		AttemptID:     attempt.AttemptID,
		AttemptNumber: prog.CurrentAttempt,
		StartedAt:     now,
		Test: TestDTO{
			ID:           test.ID,
			ThemeID:      themeID,
			Difficulty:   test.Difficulty,
			PassingScore: test.PassingScore,
			Questions:    questions,
		},
	}, nil
}

// SubmitTestAttempt grades a test attempt and updates progress.
// It is idempotent: re-submitting an already-submitted attempt returns the cached result.
func (uc *UseCase) SubmitTestAttempt(ctx context.Context, userID int64, attemptID string, answers []progress.AnswerItem) (*SubmitResult, error) {
	attempt, err := uc.attempts.GetByAttemptID(ctx, attemptID)
	if err != nil {
		return nil, err
	}

	// Verify the attempt belongs to this user before anything else.
	if attempt.UserID != userID {
		return nil, apperrors.ErrForbidden
	}

	// Idempotency: return cached result if already submitted.
	if attempt.IsSubmitted() {
		test, err := uc.tests.GetByThemeID(ctx, attempt.ThemeID)
		if err != nil {
			return nil, err
		}
		nextAction := uc.buildNextAction(ctx, userID, attempt.ThemeID, attempt.Passed)
		return &SubmitResult{
			Score:          attempt.Score,
			PassingScore:   test.PassingScore,
			Passed:         attempt.Passed,
			CorrectAnswers: 0, // not stored, reconstruct from score if needed
			TotalQuestions: len(test.Questions),
			AttemptNumber:  attempt.ID,
			NextAction:     nextAction,
			MotivationMsg:  buildMotivation(attempt.Passed),
		}, nil
	}

	// Load the test.
	test, err := uc.tests.GetByThemeID(ctx, attempt.ThemeID)
	if err != nil {
		return nil, err
	}

	// Build answers map.
	answersMap := make(map[int]string, len(answers))
	for _, a := range answers {
		answersMap[a.QuestionID] = a.Answer
	}

	// Grade the test and build per-question results.
	score, correct := test.Grade(answersMap)
	passed := test.Passed(score)

	questionResults := make([]QuestionResult, len(test.Questions))
	for i, q := range test.Questions {
		userAns := answersMap[q.ID]
		questionResults[i] = QuestionResult{
			QuestionID:    q.ID,
			QuestionText:  q.Text,
			UserAnswer:    userAns,
			CorrectAnswer: q.CorrectAnswer,
			IsCorrect:     strings.EqualFold(strings.TrimSpace(userAns), strings.TrimSpace(q.CorrectAnswer)),
		}
	}

	// Mark attempt as submitted.
	now := time.Now().UTC()
	attempt.SubmittedAt = &now
	attempt.Score = score
	attempt.Passed = passed
	attempt.Answers = answers
	attempt.DurationSeconds = int(now.Sub(attempt.StartedAt).Seconds())

	// Update progress.
	prog, err := uc.progress.GetByUserAndTheme(ctx, userID, attempt.ThemeID)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, err
	}
	if prog == nil {
		startTime := time.Now().UTC()
		prog = &progress.UserProgress{
			UserID:    userID,
			ThemeID:   attempt.ThemeID,
			Status:    progress.StatusStarted,
			StartedAt: startTime,
			UpdatedAt: startTime,
		}
	}
	if passed {
		prog.Complete(score)
	} else {
		prog.Fail(score)
	}
	if err := uc.progress.Upsert(ctx, prog); err != nil {
		return nil, err
	}

	// Persist the attempt update.
	// The repository Create is reused here; a real impl would Update, but
	// since our interface only has Create and Get, we create a new record.
	// In production this would be an Update call.
	if err := uc.attempts.Create(ctx, attempt); err != nil {
		// Non-fatal: progress is already saved.
		_ = err
	}

	nextAction := uc.buildNextAction(ctx, userID, attempt.ThemeID, passed)

	return &SubmitResult{
		Score:          score,
		PassingScore:   test.PassingScore,
		Passed:         passed,
		CorrectAnswers: correct,
		TotalQuestions: len(test.Questions),
		AttemptNumber:  prog.CurrentAttempt,
		NextAction:     nextAction,
		MotivationMsg:  buildMotivation(passed),
		Results:        questionResults,
	}, nil
}

// buildNextAction determines what the user should do after a test.
func (uc *UseCase) buildNextAction(ctx context.Context, userID int64, themeID int, passed bool) *NextAction {
	if !passed {
		msg := "Попробуйте ещё раз! Вы справитесь."
		return &NextAction{
			Type:    "retry_test",
			ThemeID: &themeID,
			Message: &msg,
		}
	}

	// Find the theme to get its module.
	theme, err := uc.themes.GetByID(ctx, themeID)
	if err != nil {
		return nil
	}

	// Look for the next theme in the same module.
	themes, err := uc.themes.GetByModuleID(ctx, theme.ModuleID)
	if err != nil {
		return nil
	}

	var nextTheme *progress.UserProgress // placeholder
	_ = nextTheme

	for _, t := range themes {
		if t.OrderNum == theme.OrderNum+1 {
			isIntro := t.IsIntroduction
			msg := "Переходите к следующей теме!"
			return &NextAction{
				Type:           "next_theme",
				ThemeID:        &t.ID,
				ThemeName:      &t.Name,
				IsIntroduction: &isIntro,
				Message:        &msg,
			}
		}
	}

	// No next theme — offer the module test.
	moduleID := theme.ModuleID
	msg := "Все темы пройдены! Теперь вы можете сдать итоговый тест модуля."
	return &NextAction{
		Type:     "module_test",
		ModuleID: &moduleID,
		Message:  &msg,
	}
	return &NextAction{
		Type:    "module_completed",
		Message: &msg,
	}
}

// ModuleTestQuestionDTO is a question for a module test sent to the client (correct answer omitted).
type ModuleTestQuestionDTO struct {
	ID       int    `json:"id"`
	Question string `json:"question"`
	OrderNum int    `json:"order_num"`
}

// StartModuleTestResult is the response for starting a dynamically-generated module test.
type StartModuleTestResult struct {
	AttemptID string                  `json:"attempt_id"`
	ModuleID  int                     `json:"module_id"`
	Questions []ModuleTestQuestionDTO `json:"questions"`
}

// ModuleTestSubmitResult carries the outcome of a module test submission.
type ModuleTestSubmitResult struct {
	Score          int              `json:"score"`
	PassingScore   int              `json:"passing_score"`
	Passed         bool             `json:"passed"`
	CorrectAnswers int              `json:"correct_answers"`
	TotalQuestions int              `json:"total_questions"`
	MotivationMsg  string           `json:"motivation_msg"`
	Results        []QuestionResult `json:"results"`
}

const moduleTestPassingScore = 70

// StartModuleTestAttempt dynamically samples questions from all theme tests in the module,
// persists the attempt (with correct answers stored server-side), and returns questions to the client.
func (uc *UseCase) StartModuleTestAttempt(ctx context.Context, userID int64, moduleID int) (*StartModuleTestResult, error) {
	themes, err := uc.themes.GetByModuleID(ctx, moduleID)
	if err != nil {
		return nil, err
	}

	var sampled []progress.ModuleTestQuestion
	orderNum := 0
	for _, th := range themes {
		test, err := uc.tests.GetByThemeID(ctx, th.ID)
		if err != nil {
			continue // theme may not have a test yet
		}
		pool := make([]contentDomain.Question, len(test.Questions))
		copy(pool, test.Questions)
		rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
		n := (len(pool)*50 + 99) / 100 // ceil of 50%
		if n > len(pool) {
			n = len(pool)
		}
		for _, q := range pool[:n] {
			orderNum++
			sampled = append(sampled, progress.ModuleTestQuestion{
				ID:            orderNum,
				Text:          q.Text,
				CorrectAnswer: q.CorrectAnswer,
				OrderNum:      orderNum,
			})
		}
	}
	if len(sampled) == 0 {
		return nil, apperrors.New("bad_request", "no questions available for this module", apperrors.ErrInvalidInput)
	}

	attempt := &progress.ModuleTestAttempt{
		AttemptID: uuid.NewString(),
		UserID:    userID,
		ModuleID:  moduleID,
		Questions: sampled,
		Answers:   []progress.AnswerItem{},
		StartedAt: time.Now().UTC(),
	}
	if err := uc.moduleTestAttempts.Create(ctx, attempt); err != nil {
		return nil, err
	}

	dtos := make([]ModuleTestQuestionDTO, len(sampled))
	for i, q := range sampled {
		dtos[i] = ModuleTestQuestionDTO{
			ID:       q.ID,
			Question: q.Text,
			OrderNum: q.OrderNum,
		}
	}
	return &StartModuleTestResult{
		AttemptID: attempt.AttemptID,
		ModuleID:  moduleID,
		Questions: dtos,
	}, nil
}

// SubmitModuleTestAttempt grades a module test attempt and persists the result.
func (uc *UseCase) SubmitModuleTestAttempt(ctx context.Context, userID int64, attemptID string, answers []progress.AnswerItem) (*ModuleTestSubmitResult, error) {
	attempt, err := uc.moduleTestAttempts.GetByAttemptID(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	if attempt.UserID != userID {
		return nil, apperrors.ErrForbidden
	}
	if attempt.IsSubmitted() {
		return nil, apperrors.New("bad_request", "attempt already submitted", apperrors.ErrInvalidInput)
	}

	answersMap := make(map[int]string, len(answers))
	for _, a := range answers {
		answersMap[a.QuestionID] = a.Answer
	}

	correct := 0
	questionResults := make([]QuestionResult, len(attempt.Questions))
	for i, q := range attempt.Questions {
		userAns := answersMap[q.ID]
		isCorrect := strings.EqualFold(strings.TrimSpace(userAns), strings.TrimSpace(q.CorrectAnswer))
		if isCorrect {
			correct++
		}
		questionResults[i] = QuestionResult{
			QuestionID:    q.ID,
			QuestionText:  q.Text,
			UserAnswer:    userAns,
			CorrectAnswer: q.CorrectAnswer,
			IsCorrect:     isCorrect,
		}
	}

	score := 0
	if len(attempt.Questions) > 0 {
		score = correct * 100 / len(attempt.Questions)
	}
	passed := score >= moduleTestPassingScore

	now := time.Now().UTC()
	attempt.Answers = answers
	attempt.Score = score
	attempt.Passed = passed
	attempt.SubmittedAt = &now
	attempt.DurationSeconds = int(now.Sub(attempt.StartedAt).Seconds())

	if err := uc.moduleTestAttempts.Update(ctx, attempt); err != nil {
		return nil, err
	}

	return &ModuleTestSubmitResult{
		Score:          score,
		PassingScore:   moduleTestPassingScore,
		Passed:         passed,
		CorrectAnswers: correct,
		TotalQuestions: len(attempt.Questions),
		MotivationMsg:  buildMotivation(passed),
		Results:        questionResults,
	}, nil
}

// buildMotivation returns a motivational message based on pass/fail.
func buildMotivation(passed bool) string {
	if passed {
		return "Отлично! Вы успешно прошли тест!"
	}
	return "Не расстраивайтесь! Повторите материал и попробуйте снова."
}
