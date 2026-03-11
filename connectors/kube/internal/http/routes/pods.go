package routes

import (
	"connector-kube/internal/http/handlers"

	"github.com/gofiber/fiber/v2"
)

const podsRouterName = "/pods"

// ListPods godoc
// @Summary      List pods
// @Description  Returns list of pods.
// @Tags         pods
// @Produce      json
// @Param        namespace               query     string  false  "Pods only from this namespace"
// @Param        container_name_contains query     string  false  "Pods only with container name matching this substring"
// @Param        status                  query     string  false  "Pods only with this status"
// @Param        status_not              query     string  false  "Pods only without this status"
// @Param        jobs_only               query     bool    false  "Pods only which are jobs"
// @Success      200  {object}  response.PodListResponse
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /pods [get]
func ListPods(router fiber.Router) {
	api := router.Group(podsRouterName)
	api.Get("/", handlers.ListPods)
}
