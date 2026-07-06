package notifications

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	asOfDate := time.Now().UTC()
	if raw := c.Query("as_of_date"); raw != "" {
		asOfDate, err = time.Parse("2006-01-02", raw)
		if err != nil {
			common.RespondError(c, common.BadRequest("as_of_date must use YYYY-MM-DD format"))
			return
		}
	}
	response, err := h.service.List(userID, asOfDate)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, response)
}
