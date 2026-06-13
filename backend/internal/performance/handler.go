package performance

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(c *gin.Context) {
	userID, portfolioID, ok := parseUserAndPortfolio(c)
	if !ok {
		return
	}

	response, err := h.service.Get(userID, portfolioID, c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, response)
}

func parseUserAndPortfolio(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		common.RespondError(c, err)
		return uuid.Nil, uuid.Nil, false
	}

	portfolioID, err := uuid.Parse(c.Param("portfolioId"))
	if err != nil {
		common.RespondError(c, common.NotFound("Portfolio not found"))
		return uuid.Nil, uuid.Nil, false
	}

	return userID, portfolioID, true
}
