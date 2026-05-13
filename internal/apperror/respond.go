package apperror

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func Respond(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.Code, gin.H{"erro": appErr.Message})
		return
	}
	// Unknown error — log it server-side (I'll add proper logging later)
	// Never send internal details to the client

	c.JSON(ErrInternal.Code, gin.H{"error": ErrInternal.Message})
}
