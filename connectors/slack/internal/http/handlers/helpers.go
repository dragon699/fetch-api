package handlers

import (
	"connector-slack/internal/config"
	"connector-slack/internal/http/dto/response"
	"errors"

	"github.com/gofiber/fiber/v2"
)

func parseBody(ctx *fiber.Ctx, dest any) error {
	if err := ctx.BodyParser(dest); err != nil {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "Invalid request payload",
			},
		)
	}

	return nil
}

func serviceError(ctx *fiber.Ctx, err error) error {
	var clientErr *config.ClientError
	if errors.As(err, &clientErr) {
		return ctx.Status(502).JSON(
			response.ErrorResponse{
				Error:            err.Error(),
				UpstreamResponse: clientErr.UpstreamResponse(),
			},
		)
	}

	return ctx.Status(500).JSON(
		response.ErrorResponse{
			Error: err.Error(),
		},
	)
}

func badRequestError(ctx *fiber.Ctx, msg string) error {
	return ctx.Status(400).JSON(
		response.ErrorResponse{
			Error: msg,
		},
	)
}
