package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/vladkonst/mnemonics/internal/domain/content"
	"github.com/vladkonst/mnemonics/internal/domain/progress"
	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/internal/domain/user"
	"github.com/vladkonst/mnemonics/internal/repository/sqlite"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

func openTestDB(t *testing.T) interface{ Close() error } {
	t.Helper()
	db, err := sqlite.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestUserRepository(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := sqlite.NewUserRepo(db)

	t.Run("Create and GetByID", func(t *testing.T) {
		u := &user.User{
			TelegramID:           12345,
			Role:                 user.RoleStudent,
			SubscriptionStatus:   user.SubscriptionStatusInactive,
			Language:             "ru",
			Timezone:             "UTC",
			NotificationsEnabled: true,
		}

		if err := repo.Create(ctx, u); err != nil {
			t.Fatalf("Create: %v", err)
		}

		got, err := repo.GetByID(ctx, 12345)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Role != user.RoleStudent {
			t.Errorf("expected role student, got %s", got.Role)
		}
	})

	t.Run("GetByID not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 99999)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperrors.IsNotFound(err) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Exists", func(t *testing.T) {
		exists, err := repo.Exists(ctx, 12345)
		if err != nil {
			t.Fatalf("Exists: %v", err)
		}
		if !exists {
			t.Error("expected user to exist")
		}

		exists, err = repo.Exists(ctx, 99999)
		if err != nil {
			t.Fatalf("Exists: %v", err)
		}
		if exists {
			t.Error("expected user not to exist")
		}
	})

	t.Run("Update", func(t *testing.T) {
		u, err := repo.GetByID(ctx, 12345)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		u.Role = user.RoleTeacher
		u.SubscriptionStatus = user.SubscriptionStatusActive

		if err := repo.Update(ctx, u); err != nil {
			t.Fatalf("Update: %v", err)
		}

		got, err := repo.GetByID(ctx, 12345)
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if got.Role != user.RoleTeacher {
			t.Errorf("expected role teacher, got %s", got.Role)
		}
	})
}

func TestModuleRepository(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := sqlite.NewModuleRepo(db)

	t.Run("Create and GetByID", func(t *testing.T) {
		desc := "Test module description"
		m := &content.Module{
			Name:        "Anatomy 101",
			Description: &desc,
			OrderNum:    1,
			IsLocked:    false,
		}

		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if m.ID == 0 {
			t.Fatal("expected module ID to be set")
		}

		got, err := repo.GetByID(ctx, m.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != "Anatomy 101" {
			t.Errorf("expected Name Anatomy 101, got %s", got.Name)
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		m2 := &content.Module{Name: "Anatomy 102", OrderNum: 2}
		if err := repo.Create(ctx, m2); err != nil {
			t.Fatalf("Create: %v", err)
		}

		modules, err := repo.GetAll(ctx)
		if err != nil {
			t.Fatalf("GetAll: %v", err)
		}
		if len(modules) < 2 {
			t.Errorf("expected at least 2 modules, got %d", len(modules))
		}
	})

	t.Run("GetByID not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 99999)
		if !apperrors.IsNotFound(err) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Update", func(t *testing.T) {
		m := &content.Module{Name: "Update Me", OrderNum: 10}
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("Create: %v", err)
		}
		m.Name = "Updated"
		if err := repo.Update(ctx, m); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, err := repo.GetByID(ctx, m.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != "Updated" {
			t.Errorf("expected Updated, got %s", got.Name)
		}
	})
}

func TestThemeRepository(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	modRepo := sqlite.NewModuleRepo(db)
	themeRepo := sqlite.NewThemeRepo(db)

	// Create parent module
	m := &content.Module{Name: "Biology", OrderNum: 1}
	if err := modRepo.Create(ctx, m); err != nil {
		t.Fatalf("Create module: %v", err)
	}

	t.Run("Create and GetByID", func(t *testing.T) {
		theme := &content.Theme{
			ModuleID:       m.ID,
			Name:           "Introduction",
			OrderNum:       1,
			IsIntroduction: true,
			IsLocked:       false,
		}
		if err := themeRepo.Create(ctx, theme); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if theme.ID == 0 {
			t.Fatal("expected theme ID to be set")
		}

		got, err := themeRepo.GetByID(ctx, theme.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if !got.IsIntroduction {
			t.Error("expected IsIntroduction to be true")
		}
	})

	t.Run("GetByModuleID", func(t *testing.T) {
		theme2 := &content.Theme{ModuleID: m.ID, Name: "Chapter 1", OrderNum: 2}
		if err := themeRepo.Create(ctx, theme2); err != nil {
			t.Fatalf("Create: %v", err)
		}

		themes, err := themeRepo.GetByModuleID(ctx, m.ID)
		if err != nil {
			t.Fatalf("GetByModuleID: %v", err)
		}
		if len(themes) < 2 {
			t.Errorf("expected at least 2 themes, got %d", len(themes))
		}
	})

	t.Run("GetPreviousTheme", func(t *testing.T) {
		// theme with order_num=2 should have previous theme with order_num=1
		themes, err := themeRepo.GetByModuleID(ctx, m.ID)
		if err != nil {
			t.Fatalf("GetByModuleID: %v", err)
		}
		if len(themes) < 2 {
			t.Skip("need at least 2 themes")
		}

		// Find the theme with order_num=2
		var themeID2 int
		for _, th := range themes {
			if th.OrderNum == 2 {
				themeID2 = th.ID
				break
			}
		}
		if themeID2 == 0 {
			t.Skip("no theme with order_num=2")
		}

		prev, err := themeRepo.GetPreviousTheme(ctx, themeID2)
		if err != nil {
			t.Fatalf("GetPreviousTheme: %v", err)
		}
		if prev.OrderNum != 1 {
			t.Errorf("expected previous theme order_num=1, got %d", prev.OrderNum)
		}
	})

	t.Run("GetPreviousTheme not found for first theme", func(t *testing.T) {
		themes, err := themeRepo.GetByModuleID(ctx, m.ID)
		if err != nil {
			t.Fatalf("GetByModuleID: %v", err)
		}
		var firstThemeID int
		for _, th := range themes {
			if th.OrderNum == 1 {
				firstThemeID = th.ID
				break
			}
		}
		_, err = themeRepo.GetPreviousTheme(ctx, firstThemeID)
		if !apperrors.IsNotFound(err) {
			t.Errorf("expected ErrNotFound for first theme, got %v", err)
		}
	})
}

func TestProgressRepository(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Setup: create user, module, theme
	userRepo := sqlite.NewUserRepo(db)
	modRepo := sqlite.NewModuleRepo(db)
	themeRepo := sqlite.NewThemeRepo(db)
	progressRepo := sqlite.NewProgressRepo(db)

	u := &user.User{
		TelegramID: 111,
		Role: user.RoleStudent, SubscriptionStatus: user.SubscriptionStatusInactive,
		Language: "ru", Timezone: "UTC", NotificationsEnabled: true,
	}
	if err := userRepo.Create(ctx, u); err != nil {
		t.Fatalf("Create user: %v", err)
	}

	m := &content.Module{Name: "Test Module", OrderNum: 1}
	if err := modRepo.Create(ctx, m); err != nil {
		t.Fatalf("Create module: %v", err)
	}

	theme := &content.Theme{ModuleID: m.ID, Name: "Theme 1", OrderNum: 1}
	if err := themeRepo.Create(ctx, theme); err != nil {
		t.Fatalf("Create theme: %v", err)
	}

	t.Run("Upsert and GetByUserAndTheme", func(t *testing.T) {
		p := &progress.UserProgress{
			UserID:    u.TelegramID,
			ThemeID:   theme.ID,
			Status:    progress.StatusStarted,
			StartedAt: time.Now().UTC(),
		}

		if err := progressRepo.Upsert(ctx, p); err != nil {
			t.Fatalf("Upsert: %v", err)
		}

		got, err := progressRepo.GetByUserAndTheme(ctx, u.TelegramID, theme.ID)
		if err != nil {
			t.Fatalf("GetByUserAndTheme: %v", err)
		}
		if got.Status != progress.StatusStarted {
			t.Errorf("expected StatusStarted, got %s", got.Status)
		}

		// Update via upsert
		score := 85
		p.Status = progress.StatusCompleted
		p.Score = &score
		now := time.Now().UTC()
		p.CompletedAt = &now

		if err := progressRepo.Upsert(ctx, p); err != nil {
			t.Fatalf("Upsert update: %v", err)
		}

		got, err = progressRepo.GetByUserAndTheme(ctx, u.TelegramID, theme.ID)
		if err != nil {
			t.Fatalf("GetByUserAndTheme after update: %v", err)
		}
		if got.Status != progress.StatusCompleted {
			t.Errorf("expected StatusCompleted, got %s", got.Status)
		}
		if got.Score == nil || *got.Score != 85 {
			t.Errorf("expected score 85, got %v", got.Score)
		}
	})

	t.Run("GetByUserAndTheme not found", func(t *testing.T) {
		_, err := progressRepo.GetByUserAndTheme(ctx, 99999, 99999)
		if !apperrors.IsNotFound(err) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("GetByUser", func(t *testing.T) {
		result, err := progressRepo.GetByUser(ctx, u.TelegramID)
		if err != nil {
			t.Fatalf("GetByUser: %v", err)
		}
		if len(result) == 0 {
			t.Error("expected at least one progress record")
		}
	})

	t.Run("CountCompletedByUser", func(t *testing.T) {
		count, err := progressRepo.CountCompletedByUser(ctx, u.TelegramID)
		if err != nil {
			t.Fatalf("CountCompletedByUser: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 completed, got %d", count)
		}
	})
}

func TestPromoCodeRepository(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Create teacher user first (FK constraint)
	userRepo := sqlite.NewUserRepo(db)
	teacher := &user.User{
		TelegramID: 999,
		Role: user.RoleTeacher, SubscriptionStatus: user.SubscriptionStatusInactive,
		Language: "ru", Timezone: "UTC", NotificationsEnabled: true,
	}
	if err := userRepo.Create(ctx, teacher); err != nil {
		t.Fatalf("Create teacher: %v", err)
	}

	repo := sqlite.NewPromoCodeRepo(db)

	t.Run("Create and GetByCode", func(t *testing.T) {
		p := &subscription.PromoCode{
			Code:           "TEST123",
			UniversityName: "Test University",
			MaxActivations: 10,
			Remaining:      10,
			Status:         subscription.PromoCodeStatusPending,
		}

		if err := repo.Create(ctx, p); err != nil {
			t.Fatalf("Create: %v", err)
		}

		got, err := repo.GetByCode(ctx, "TEST123")
		if err != nil {
			t.Fatalf("GetByCode: %v", err)
		}
		if got.UniversityName != "Test University" {
			t.Errorf("expected Test University, got %s", got.UniversityName)
		}
		if got.Status != subscription.PromoCodeStatusPending {
			t.Errorf("expected pending status, got %s", got.Status)
		}
	})

	t.Run("Activate (Activate domain method + Update)", func(t *testing.T) {
		got, err := repo.GetByCode(ctx, "TEST123")
		if err != nil {
			t.Fatalf("GetByCode: %v", err)
		}

		if err := got.Activate(teacher.TelegramID); err != nil {
			t.Fatalf("domain Activate: %v", err)
		}

		if err := repo.Update(ctx, got); err != nil {
			t.Fatalf("Update: %v", err)
		}

		updated, err := repo.GetByCode(ctx, "TEST123")
		if err != nil {
			t.Fatalf("GetByCode: %v", err)
		}
		if updated.Status != subscription.PromoCodeStatusActive {
			t.Errorf("expected active status, got %s", updated.Status)
		}
		if updated.TeacherID == nil || *updated.TeacherID != teacher.TelegramID {
			t.Errorf("expected teacher ID %d", teacher.TelegramID)
		}
	})

	t.Run("GetByTeacherID", func(t *testing.T) {
		codes, err := repo.GetByTeacherID(ctx, teacher.TelegramID)
		if err != nil {
			t.Fatalf("GetByTeacherID: %v", err)
		}
		if len(codes) == 0 {
			t.Error("expected at least one promo code")
		}
	})

	t.Run("Deactivate", func(t *testing.T) {
		if err := repo.Deactivate(ctx, "TEST123"); err != nil {
			t.Fatalf("Deactivate: %v", err)
		}

		got, err := repo.GetByCode(ctx, "TEST123")
		if err != nil {
			t.Fatalf("GetByCode after deactivate: %v", err)
		}
		if got.Status != subscription.PromoCodeStatusDeactivated {
			t.Errorf("expected deactivated status, got %s", got.Status)
		}
	})

	t.Run("GetByCode not found", func(t *testing.T) {
		_, err := repo.GetByCode(ctx, "NONEXISTENT")
		if !apperrors.IsNotFound(err) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
