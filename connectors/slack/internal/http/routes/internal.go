package routes

import (
	_ "connector-slack/internal/http/dto/response"
	"connector-slack/internal/http/handlers"

	"github.com/gofiber/fiber/v2"
)

const apiRouterName = "/api"

// Health godoc
// @Summary      Health check
// @Tags         health
// @Produce      json
// @Success      200  {object}  response.HealthResponse
// @Router       /api/health [get]
func Health(router fiber.Router) {
	api := router.Group(apiRouterName)
	api.Get("/health", handlers.Health)
}

// Ready godoc
// @Summary      Readiness check
// @Tags         health
// @Produce      json
// @Success      200  {object}  response.ReadyResponse
// @Router       /api/ready [get]
func Ready(router fiber.Router) {
	api := router.Group(apiRouterName)
	api.Get("/ready", handlers.Ready)
}
