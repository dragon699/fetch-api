package handlers

import (
	"connector-kube/internal/config"
	"connector-kube/internal/dto/response"

	"github.com/gofiber/fiber/v2"
)

func Health(ctx *fiber.Ctx) error {
	return ctx.JSON(
		response.HealthResponse{
			ConnectorName:   config.Config.Name,
			Healthy:         config.Config.Healthy,
			HealthLastCheck: *config.Config.HealthLastCheck,
			HealthNextCheck: *config.Config.HealthNextCheck,
		},
	)
}

func Ready(ctx *fiber.Ctx) error {
	return ctx.JSON(
		response.ReadyResponse{
			Ready: config.Config.Healthy,
		},
	)
}
