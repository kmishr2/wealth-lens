package contributions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/middleware"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }
func (h *Handler) Get(c *gin.Context) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	portfolioID, err := uuid.Parse(c.Param("portfolioId"))
	if err != nil {
		common.RespondError(c, common.NotFound("Portfolio not found"))
		return
	}
	response, err := h.service.Get(userID, portfolioID, c.Query("start_date"), c.Query("end_date"), c.Query("currency"))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, response)
}
