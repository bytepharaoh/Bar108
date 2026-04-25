package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// parseID extracts and parses the :id URL param.
// Returns (id, true) on success.
// Writes a 400 response and returns (0, false) on failure.
// Shared across all handlers — defined once here instead of
// repeating the same 4 lines in every handler method.
func parseID(c *gin.Context) (int32, bool) {
	id64, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return int32(id64), true
}
