package routes

import (
	"crypto/subtle"

	"connector-downloader/internal/config"
	"connector-downloader/internal/mcp"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

// MCP godoc
// @Summary      MCP server (Streamable HTTP)
// @Description  Model Context Protocol endpoint exposing the search_torrents, download, list_downloads and get_download tools.
// @Description  Stateless JSON-RPC over POST with plain JSON responses. Only registered when MCP_AUTH_TOKEN is set.
// @Tags         mcp
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer <MCP_AUTH_TOKEN>"
// @Param        request        body      object  true  "JSON-RPC 2.0 request, e.g. initialize, tools/list or tools/call"
// @Success      200            {object}  object  "JSON-RPC 2.0 response"
// @Failure      401            "Missing or invalid bearer token"
// @Router       /mcp [post]
func MCP(router fiber.Router) {
	if config.Config.McpAuthToken == "" {
		return
	}

	expected := []byte("Bearer " + config.Config.McpAuthToken)

	auth := func(ctx *fiber.Ctx) error {
		if subtle.ConstantTimeCompare([]byte(ctx.Get(fiber.HeaderAuthorization)), expected) != 1 {
			return ctx.SendStatus(fiber.StatusUnauthorized)
		}

		return ctx.Next()
	}

	router.All("/mcp", auth, adaptor.HTTPHandler(mcp.NewHandler(config.Config.OtelServiceVersion)))
}
