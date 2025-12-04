package user

// ========== Create User ==========

type CreateUserInput struct {
	Email    string
	Username string
	Password string
	Name     string
	Role     string
}

// ========== Update User ==========

type UpdateUserInput struct {
	Id       int
	Email    *string
	Username *string
	Name     *string
	Role     *string
	IsActive *bool
}

// ========== Change Password ==========

type ChangePasswordInput struct {
	UserId          int
	CurrentPassword string
	NewPassword     string
}

// ========== Get Users ==========

type GetUsersInput struct {
	Page       int
	Size       int
	OnlyActive bool
}

// ========== Login ==========

type LoginInput struct {
	Email    string
	Password string
}

