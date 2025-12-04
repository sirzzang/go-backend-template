package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/your-org/go-backend-template/internal/app/server/handler"
	"github.com/your-org/go-backend-template/internal/app/server/service/user"
)

// Handler handles user-related HTTP requests.
type Handler struct {
	handler.BaseHandler
	userService *user.Service
	jwtService  IJWTService
}

// IJWTService defines the interface for JWT operations.
type IJWTService interface {
	GenerateToken(userId int, role string) (string, error)
}

// NewHandler creates a new user handler.
func NewHandler(userService *user.Service, jwtService IJWTService) *Handler {
	return &Handler{
		BaseHandler: handler.BaseHandler{},
		userService: userService,
		jwtService:  jwtService,
	}
}

// CreateUser handles POST /users
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleBindingError(c, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.HandleValidationError(c, err.Error())
		return
	}

	input := &user.CreateUserInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
		Name:     req.Name,
		Role:     req.Role,
	}

	userId, err := h.userService.CreateUser(input)
	if err != nil {
		h.HandleDomainError(c, err)
		return
	}

	h.HandleSuccess(c, http.StatusCreated, &CreateUserResponse{Id: userId})
}

// GetUser handles GET /users/:id
func (h *Handler) GetUser(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.HandleValidationError(c, "invalid user id")
		return
	}

	gotUser, err := h.userService.GetUserById(userId)
	if err != nil {
		h.HandleDomainError(c, err)
		return
	}

	h.HandleSuccess(c, http.StatusOK, ToUserResponse(gotUser))
}

// GetUsers handles GET /users
func (h *Handler) GetUsers(c *gin.Context) {
	var query GetUsersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.HandleBindingError(c, err)
		return
	}

	input := &user.GetUsersInput{
		Page:       query.GetPage(),
		Size:       query.GetSize(),
		OnlyActive: query.GetOnlyActive(),
	}

	result, err := h.userService.GetUsers(input)
	if err != nil {
		h.HandleDomainError(c, err)
		return
	}

	resp := &GetUsersResponse{
		TotalCount: result.TotalCount,
		Count:      len(result.Users),
		Data:       ToUserResponseList(result.Users),
	}

	h.HandleSuccess(c, http.StatusOK, resp)
}

// UpdateUser handles PATCH /users/:id
func (h *Handler) UpdateUser(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.HandleValidationError(c, "invalid user id")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleBindingError(c, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.HandleValidationError(c, err.Error())
		return
	}

	input := &user.UpdateUserInput{
		Id:       userId,
		Email:    req.Email,
		Username: req.Username,
		Name:     req.Name,
		Role:     req.Role,
		IsActive: req.IsActive,
	}

	if err := h.userService.UpdateUser(input); err != nil {
		h.HandleDomainError(c, err)
		return
	}

	h.HandleSuccess(c, http.StatusOK, &MessageResponse{Message: "user updated successfully"})
}

// DeleteUser handles DELETE /users/:id
func (h *Handler) DeleteUser(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.HandleValidationError(c, "invalid user id")
		return
	}

	if err := h.userService.DeleteUser(userId); err != nil {
		h.HandleDomainError(c, err)
		return
	}

	h.HandleSuccess(c, http.StatusOK, &MessageResponse{Message: "user deleted successfully"})
}

// ChangePassword handles POST /users/:id/change-password
func (h *Handler) ChangePassword(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.HandleValidationError(c, "invalid user id")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleBindingError(c, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.HandleValidationError(c, err.Error())
		return
	}

	input := &user.ChangePasswordInput{
		UserId:          userId,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	if err := h.userService.ChangePassword(input); err != nil {
		h.HandleDomainError(c, err)
		return
	}

	h.HandleSuccess(c, http.StatusOK, &MessageResponse{Message: "password changed successfully"})
}

// GetMe handles GET /me
func (h *Handler) GetMe(c *gin.Context) {
	userId := handler.GetUserId(c)

	gotUser, err := h.userService.GetUserById(userId)
	if err != nil {
		h.HandleDomainError(c, err)
		return
	}

	h.HandleSuccess(c, http.StatusOK, ToUserResponse(gotUser))
}

// Login handles POST /auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleBindingError(c, err)
		return
	}

	input := &user.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	loggedInUser, err := h.userService.Login(input)
	if err != nil {
		h.HandleDomainError(c, err)
		return
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(loggedInUser.Id, loggedInUser.Role)
	if err != nil {
		h.HandleDomainError(c, err)
		return
	}

	resp := &LoginResponse{
		Token: token,
		User:  ToUserResponse(loggedInUser),
	}

	h.HandleSuccess(c, http.StatusOK, resp)
}

