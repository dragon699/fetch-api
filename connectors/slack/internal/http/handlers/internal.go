package handlers

import (
	"connector-slack/internal/config"
	"connector-slack/internal/http/dto/response"

	"github.com/gofiber/fiber/v2"
)

func Health(ctx *fiber.Ctx) error {
	lastCheck := ""
	nextCheck := ""

	if config.Config.HealthLastCheck != nil {
		lastCheck = *config.Config.HealthLastCheck
	}

	if config.Config.HealthNextCheck != nil {
		nextCheck = *config.Config.HealthNextCheck
	}

	return ctx.JSON(
		response.HealthResponse{
			ConnectorName:   config.Config.Name,
			Healthy:         config.Config.Healthy,
			HealthLastCheck: lastCheck,
			HealthNextCheck: nextCheck,
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
