package user

import "time"

// User is the aggregate root for the user domain.
type User struct {
	TelegramID           int64
	FirstName            string
	LastName             *string
	Username             *string
	Role                 Role
	SubscriptionStatus   SubscriptionStatus
	UniversityCode       *string
	PendingPaymentID     *string
	Language             string
	Timezone             string
	NotificationsEnabled bool
	LastActivityAt       *time.Time
	CreatedAt            time.Time
}

// IsTeacher returns true if the user has the teacher role.
func (u *User) IsTeacher() bool {
	return u.Role == RoleTeacher
}

// HasActiveSubscription returns true if the user's subscription is active.
func (u *User) HasActiveSubscription() bool {
	return u.SubscriptionStatus == SubscriptionStatusActive
}

// SetRole updates the user role.
func (u *User) SetRole(r Role) {
	u.Role = r
}

// ActivateSubscription marks the user's subscription as active.
func (u *User) ActivateSubscription(universityCode *string) {
	u.SubscriptionStatus = SubscriptionStatusActive
	if universityCode != nil {
		u.UniversityCode = universityCode
	}
}

// DeactivateSubscription marks the user's subscription as expired.
func (u *User) DeactivateSubscription() {
	u.SubscriptionStatus = SubscriptionStatusExpired
}

// SetPendingPayment records a pending payment ID on the user.
func (u *User) SetPendingPayment(paymentID string) {
	u.PendingPaymentID = &paymentID
}

// ClearPendingPayment removes the pending payment ID.
func (u *User) ClearPendingPayment() {
	u.PendingPaymentID = nil
}
