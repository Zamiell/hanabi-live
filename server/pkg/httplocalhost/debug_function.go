package httplocalhost

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Manager) debugFunction(c *gin.Context) {
	// TODO
	// debugFunction(c)
	c.String(http.StatusOK, "success\n")
}
