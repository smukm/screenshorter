package handlers

import (
	"screenshorter/internal/middleware"
	"screenshorter/internal/service"

	gin "github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {

	router := gin.New()

	router.Use()

	api := router.Group("/api")
	api.Use(middleware.BearerAuthMiddleware())
	{
		api.POST("gz", h.Make)
	}

	return router
}
