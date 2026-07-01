package benchmarks

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
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		common.RespondError(c, err)
		return
	}

	var req BenchmarkCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	benchmark, err := h.service.Create(userID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusCreated, benchmark)
}

func (h *Handler) List(c *gin.Context) {
	benchmarks, err := h.service.List(common.ParsePagination(c))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, benchmarks)
}

func (h *Handler) CreateObservation(c *gin.Context) {
	userID, benchmarkID, ok := parseUserAndBenchmark(c)
	if !ok {
		return
	}

	var req BenchmarkObservationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, common.BadRequest("Invalid request body"))
		return
	}

	observation, err := h.service.CreateObservation(userID, benchmarkID, req)
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusCreated, observation)
}

func (h *Handler) ListObservations(c *gin.Context) {
	benchmarkID, ok := parseBenchmark(c)
	if !ok {
		return
	}

	observations, err := h.service.ListObservations(benchmarkID, common.ParsePagination(c))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, observations)
}

func (h *Handler) ComparePortfolio(c *gin.Context) {
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

	benchmarkID, ok := parseBenchmark(c)
	if !ok {
		return
	}

	response, err := h.service.ComparePortfolio(userID, portfolioID, benchmarkID, c.Query("start_date"), c.Query("end_date"), c.Query("currency"))
	if err != nil {
		common.RespondError(c, err)
		return
	}
	common.RespondOK(c, http.StatusOK, response)
}

func parseUserAndBenchmark(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	userID, err := middleware.CurrentUserID(c)
	if err != nil {
		common.RespondError(c, err)
		return uuid.Nil, uuid.Nil, false
	}

	benchmarkID, ok := parseBenchmark(c)
	if !ok {
		return uuid.Nil, uuid.Nil, false
	}

	return userID, benchmarkID, true
}

func parseBenchmark(c *gin.Context) (uuid.UUID, bool) {
	benchmarkID, err := uuid.Parse(c.Param("benchmarkId"))
	if err != nil {
		common.RespondError(c, common.NotFound("Benchmark not found"))
		return uuid.Nil, false
	}

	return benchmarkID, true
}
