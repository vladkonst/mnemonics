package user

import "fmt"

// Role is a value object representing the user's role.
type Role string

const (
	RoleStudent Role = "student"
	RoleTeacher Role = "teacher"
)

// NewRole creates a Role from a string, returning an error for invalid values.
func NewRole(s string) (Role, error) {
	switch Role(s) {
	case RoleStudent, RoleTeacher:
		return Role(s), nil
	default:
		return "", fmt.Errorf("invalid role %q: must be student or teacher", s)
	}
}

func (r Role) String() string { return string(r) }
