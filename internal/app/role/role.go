package role

type Role int

const (
	Guest Role = iota
	User
	Admin
)

func FromString(s string) Role {
	switch s {
	case "admin", "Admin":
		return Admin
	case "user", "User":
		return User
	default:
		return Guest
	}
}
