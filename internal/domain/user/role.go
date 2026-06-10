package user

import "fmt"

// Role is a value object representing the user's role.
type Role string

const (
	RoleUnknown  Role = "unknown"
	RoleStudent  Role = "student"
	RoleTeacher  Role = "teacher"
	RoleManager  Role = "manager"
)

// NewRole creates a Role from a string, returning an error for invalid values.
func NewRole(s string) (Role, error) {
	switch Role(s) {
	case RoleUnknown, RoleStudent, RoleTeacher, RoleManager:
		return Role(s), nil
	default:
		return "", fmt.Errorf("invalid role %q: must be unknown, student, teacher or manager", s)
	}
}

func (r Role) String() string { return string(r) }
