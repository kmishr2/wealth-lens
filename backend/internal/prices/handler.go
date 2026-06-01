package prices

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

func (h *Handler) Create(c *gin.Context) {
	userID, assetID, ok := parseUserAndAsset(c)
	if !ok {
		return
	}

	var req AssetPriceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	price, err := h.service.Create(userID, assetID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusCreated, price)
}

func (h *Handler) List(c *gin.Context) {
	assetID, ok := parseAsset(c)
	if !ok {
		return
	}

	prices, err := h.service.List(assetID, common.ParsePagination(c))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, prices)
}

func (h *Handler) GetLatest(c *gin.Context) {
	assetID, ok := parseAsset(c)
	if !ok {
		return
	}

	price, err := h.service.GetLatest(assetID)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, price)
}

func parseUserAndAsset(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		common.RespondError(c, err)
		return uuid.Nil, uuid.Nil, false
	}

	assetID, ok := parseAsset(c)
	if !ok {
		return uuid.Nil, uuid.Nil, false
	}

	return userID, assetID, true
}

func parseAsset(c *gin.Context) (uuid.UUID, bool) {
	assetID, err := uuid.Parse(c.Param("assetId"))
	if err != nil {
		common.RespondError(c, common.NotFound("Asset not found"))
		return uuid.Nil, false
	}

	return assetID, true
}
