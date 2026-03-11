package handlers

import (
	"common/utils"
	"connector-kube/internal/dto/response"
	"connector-kube/internal/kubernetes"

	"github.com/gofiber/fiber/v2"
)

func ListNamespaces(ctx *fiber.Ctx) error {
	namespaces, err := kubernetes.Client.ListNamespaces()

	if err != nil {
		return ctx.Status(500).JSON(
			response.ErrorResponse{
				Error: err.Error(),
			},
		)
	}

	result := make([]response.Namespace, 0, len(namespaces))

	for _, namespace := range namespaces {
		namespaceData := response.Namespace{
			Name:      namespace.ObjectMeta.Name,
			CreatedAt: utils.TrimKubeTime(namespace.ObjectMeta.CreationTimestamp),
		}

		result = append(result, namespaceData)
	}

	return ctx.JSON(
		response.BaseResponse[response.Namespace]{
			TotalItems: len(result),
			Items:      result,
		},
	)
}
