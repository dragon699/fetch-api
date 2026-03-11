package routes

import (
	"connector-kube/internal/http/handlers"

	"github.com/gofiber/fiber/v2"
)

const namespacesRouterName = "/namespaces"

// ListNamespaces
// @Summary      List namespaces
// @Description  Returns list of Namespaces.
// @Tags         namespaces
// @Produce      json
// @Success      200  {object}  response.NamespaceListResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /namespaces [get]
func ListNamespaces(router fiber.Router) {
	api := router.Group(namespacesRouterName)
	api.Get("/", handlers.ListNamespaces)
}
