package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/your-org/go-backend-template/internal/pkg/domain"
)

// BaseHandler provides common response and error handling methods.
// Embed this in domain-specific handlers to reuse common functionality.
type BaseHandler struct{}

// HandleBindingError handles request binding errors (gin ShouldBind fails).
func (b *BaseHandler) HandleBindingError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		messages := make([]string, 0, len(validationErrs))
		for _, e := range validationErrs {
			messages = append(messages, e.Error())
		}
		b.responseError(c, http.StatusBadRequest, "invalid request format", messages)
		return
	}

	b.responseError(c, http.StatusBadRequest, "invalid request format", err.Error())
}

// HandleValidationError handles DTO validation errors (req.IsValid() fails).
func (b *BaseHandler) HandleValidationError(c *gin.Context, message string) {
	b.responseError(c, http.StatusBadRequest, message)
}

// HandleDomainError handles domain errors returned from service layer.
func (b *BaseHandler) HandleDomainError(c *gin.Context, err error) {
	if domainErr, ok := err.(domain.DomainError); ok {
		b.responseError(c, domainErr.HTTPStatus(), domainErr.Error())
		return
	}

	// Default to internal server error for unknown errors
	b.responseError(c, http.StatusInternalServerError, err.Error())
}

// HandleSuccess sends a success response.
func (b *BaseHandler) HandleSuccess(c *gin.Context, status int, data ...any) {
	if len(data) > 0 {
		c.JSON(status, data[0])
	} else {
		c.Status(status)
	}
}

// responseError sends an error response with the given status code and message.
func (b *BaseHandler) responseError(c *gin.Context, code int, msg string, data ...any) {
	response := gin.H{"message": msg}
	if len(data) > 0 {
		response["data"] = data[0]
	}
	c.AbortWithStatusJSON(code, response)
}
