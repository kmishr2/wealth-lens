package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/accounts"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/allocations"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/auth"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/benchmarks"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/beta"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/config"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/goals"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/health"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/holdings"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/middleware"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/performance"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/risk"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/snapshots"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/transactions"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/users"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/valuations"
	"gorm.io/gorm"
)

func New(cfg config.Config, db *gorm.DB) (*gin.Engine, error) {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, err
	}
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	userRepo := users.NewRepository(db)
	authRepo := auth.NewRepository(db)
	portfolioRepo := portfolios.NewRepository(db)
	accountRepo := accounts.NewRepository(db)
	assetRepo := assets.NewRepository(db)
	transactionRepo := transactions.NewRepository(db)
	holdingsRepo := holdings.NewRepository(db)
	priceRepo := prices.NewRepository(db)
	snapshotRepo := snapshots.NewRepository(db)
	benchmarkRepo := benchmarks.NewRepository(db)
	goalRepo := goals.NewRepository(db)

	authService := auth.NewService(cfg, authRepo, userRepo)
	userService := users.NewService(userRepo)
	portfolioService := portfolios.NewService(portfolioRepo)
	accountService := accounts.NewService(accountRepo, portfolioRepo)
	assetService := assets.NewService(assetRepo)
	transactionService := transactions.NewService(transactionRepo, portfolioRepo, accountRepo, assetRepo)
	holdingsService := holdings.NewService(holdingsRepo, portfolioRepo)
	priceService := prices.NewService(priceRepo, assetRepo)
	valuationService := valuations.NewService(holdingsRepo, priceRepo, portfolioRepo)
	allocationService := allocations.NewService(holdingsRepo, priceRepo, portfolioRepo)
	snapshotService := snapshots.NewService(snapshotRepo, holdingsRepo, priceRepo, portfolioRepo)
	performanceService := performance.NewService(portfolioRepo, snapshotRepo, transactionRepo)
	riskService := risk.NewService(portfolioRepo, snapshotRepo, transactionRepo)
	benchmarkService := benchmarks.NewService(benchmarkRepo, portfolioRepo, snapshotRepo)
	goalService := goals.NewService(goalRepo, portfolioRepo, snapshotRepo)
	healthService := health.NewService(allocationService, riskService)
	betaService := beta.NewService(portfolioRepo, benchmarkRepo, snapshotRepo, transactionRepo)

	authHandler := auth.NewHandler(authService)
	userHandler := users.NewHandler(userService)
	portfolioHandler := portfolios.NewHandler(portfolioService)
	accountHandler := accounts.NewHandler(accountService)
	assetHandler := assets.NewHandler(assetService)
	transactionHandler := transactions.NewHandler(transactionService)
	holdingsHandler := holdings.NewHandler(holdingsService)
	priceHandler := prices.NewHandler(priceService)
	valuationHandler := valuations.NewHandler(valuationService)
	allocationHandler := allocations.NewHandler(allocationService)
	snapshotHandler := snapshots.NewHandler(snapshotService)
	performanceHandler := performance.NewHandler(performanceService)
	riskHandler := risk.NewHandler(riskService)
	benchmarkHandler := benchmarks.NewHandler(benchmarkService)
	goalHandler := goals.NewHandler(goalService)
	healthHandler := health.NewHandler(healthService)
	betaHandler := beta.NewHandler(betaService)

	v1 := router.Group("/api/v1")
	authRoutes := v1.Group("")
	authRoutes.Use(middleware.RateLimit(cfg.AuthRateLimit, cfg.AuthRateWindow))
	auth.RegisterRoutes(authRoutes, authHandler)

	protected := v1.Group("")
	protected.Use(middleware.RequireAuth(authService))
	users.RegisterRoutes(protected, userHandler)
	portfolios.RegisterRoutes(protected, portfolioHandler)
	accounts.RegisterRoutes(protected, accountHandler)
	assets.RegisterRoutes(protected, assetHandler)
	transactions.RegisterRoutes(protected, transactionHandler)
	holdings.RegisterRoutes(protected, holdingsHandler)
	prices.RegisterRoutes(protected, priceHandler)
	valuations.RegisterRoutes(protected, valuationHandler)
	allocations.RegisterRoutes(protected, allocationHandler)
	snapshots.RegisterRoutes(protected, snapshotHandler)
	performance.RegisterRoutes(protected, performanceHandler)
	risk.RegisterRoutes(protected, riskHandler)
	benchmarks.RegisterRoutes(protected, benchmarkHandler)
	goals.RegisterRoutes(protected, goalHandler)
	health.RegisterRoutes(protected, healthHandler)
	beta.RegisterRoutes(protected, betaHandler)

	return router, nil
}
