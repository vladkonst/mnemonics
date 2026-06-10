package subscription

import "time"

// InviteLink is a unique shareable link for a corporate student group.
// MaxActivations caps how many students may join via this link.
type InviteLink struct {
	ID             string    `json:"id"`
	TeacherID      int64     `json:"teacher_id"`
	MaxActivations int       `json:"max_activations"`
	CreatedAt      time.Time `json:"created_at"`
}

// InviteLinkActivation records a student who joined via an invite link.
type InviteLinkActivation struct {
	ID           int       `json:"id"`
	InviteLinkID string    `json:"invite_link_id"`
	UserID       int64     `json:"user_id"`
	ActivatedAt  time.Time `json:"activated_at"`
}

// InviteLinkWithStats enriches an InviteLink with aggregated activation data.
type InviteLinkWithStats struct {
	InviteLink
	ActivationCount int                    `json:"activation_count"`
	Activations     []*InviteLinkActivation `json:"activations,omitempty"`
}
