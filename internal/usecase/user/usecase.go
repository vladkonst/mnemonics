// Package user provides use cases for user management.
package user

import (
	"context"
	"time"

	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
	"github.com/vladkonst/mnemonics/internal/domain/subscription"
	"github.com/vladkonst/mnemonics/internal/domain/user"
	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// UseCase orchestrates user-related business operations.
type UseCase struct {
	users         interfaces.UserRepository
	subscriptions interfaces.SubscriptionRepository
}

// NewUseCase creates a new user UseCase.
func NewUseCase(users interfaces.UserRepository, subscriptions interfaces.SubscriptionRepository) *UseCase {
	return &UseCase{
		users:         users,
		subscriptions: subscriptions,
	}
}

// Register creates a new user if they do not already exist.
// Returns ErrAlreadyExists if the user is already registered.
func (uc *UseCase) Register(ctx context.Context, telegramID int64, firstName, lastName, username string) (*user.User, error) {
	exists, err := uc.users.Exists(ctx, telegramID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrAlreadyExists
	}

	now := time.Now().UTC()
	var lastNamePtr *string
	if lastName != "" {
		lastNamePtr = &lastName
	}
	var usernamePtr *string
	if username != "" {
		usernamePtr = &username
	}

	u := &user.User{
		TelegramID:           telegramID,
		FirstName:            firstName,
		LastName:             lastNamePtr,
		Username:             usernamePtr,
		Role:                 user.RoleStudent,
		SubscriptionStatus:   user.SubscriptionStatusInactive,
		Language:             "ru",
		Timezone:             "UTC",
		NotificationsEnabled: true,
		CreatedAt:            now,
	}

	if err := uc.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateRole changes the user's role.
func (uc *UseCase) UpdateRole(ctx context.Context, telegramID int64, role user.Role) (*user.User, error) {
	u, err := uc.users.GetByID(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	u.SetRole(role)
	if err := uc.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateSettings updates optional user settings.
func (uc *UseCase) UpdateSettings(ctx context.Context, telegramID int64, language *string, notificationsEnabled *bool) (*user.User, error) {
	u, err := uc.users.GetByID(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	if language != nil {
		u.Language = *language
	}
	if notificationsEnabled != nil {
		u.NotificationsEnabled = *notificationsEnabled
	}

	if err := uc.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// GetSubscription returns the active subscription for a user, or ErrNotFound if none.
func (uc *UseCase) GetSubscription(ctx context.Context, userID int64) (*subscription.Subscription, error) {
	// Ensure user exists.
	if _, err := uc.users.GetByID(ctx, userID); err != nil {
		return nil, err
	}

	sub, err := uc.subscriptions.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return sub, nil
}
