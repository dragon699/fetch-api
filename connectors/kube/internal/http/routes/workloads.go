package routes

import (
	"connector-kube/internal/http/handlers"

	"github.com/gofiber/fiber/v2"
)

const workloadsRouterName = "/workloads"

// ListDeployments
// @Summary      List deployments
// @Description  Returns list of Deployments.
// @Tags         workloads
// @Produce      json
// @Param        namespace               query     string  false  "Deployments only from this namespace"
// @Success      200  {object}  response.DeploymentListResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /workloads/deployments [get]
func ListDeployments(router fiber.Router) {
	api := router.Group(workloadsRouterName)
	api.Get("/deployments", handlers.ListWorkloads("deployments"))
}
