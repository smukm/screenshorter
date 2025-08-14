package handlers

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"screenshoter/config"
	"screenshoter/internal/middleware"
	"screenshoter/internal/service"

	gin "github.com/gin-gonic/gin"
)

type Handler struct {
	service    *service.Service
	cfg        *config.Config
	workerPool chan struct{} // Семафор для ограничения одновременных запросов
}

func NewHandler(service *service.Service, cfg *config.Config) *Handler {
	return &Handler{
		service:    service,
		cfg:        cfg,
		workerPool: make(chan struct{}, cfg.MaxWorkers), // Пулинг воркеров
	}
}

func (h *Handler) InitRoutes() *gin.Engine {

	router := gin.New()

	router.Use()

	api := router.Group("/api")
	api.Use(middleware.BearerAuthMiddleware(h.cfg))
	{
		api.POST("screen", h.Make)
	}

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/worker-stats", h.MetricsHandler)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return router
}
