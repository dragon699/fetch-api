package handlers

import (
	"sort"

	"common/utils"
	"connector-kube/internal/dto/request"
	"connector-kube/internal/dto/response"
	"connector-kube/internal/kubernetes"

	"github.com/gofiber/fiber/v2"
	// appsv1 "k8s.io/api/apps/v1"
)

func ListWorkloads(kind string) fiber.Handler {
	if kind == "deployments" {
		return listDeployments
	}

	return listDeployments
}

func listDeployments(ctx *fiber.Ctx) error {
	var params request.ListDeploymentsParams

	if err := ctx.QueryParser(&params); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{
				Error: err.Error(),
			},
		)
	}

	deployments, err := kubernetes.Client.ListDeployments()

	if err != nil {
		return ctx.Status(500).JSON(
			response.ErrorResponse{
				Error: err.Error(),
			},
		)
	}

	result := make([]response.Deployment, 0, len(deployments))

	for _, deployment := range deployments {
		if params.Namespace != "" && deployment.ObjectMeta.Namespace != params.Namespace {
			continue
		}

		deploymentData := response.Deployment{
			Name:      deployment.ObjectMeta.Name,
			Namespace: deployment.ObjectMeta.Namespace,
			CreatedAt: utils.TrimKubeTime(deployment.ObjectMeta.CreationTimestamp),
			Replicas:  *deployment.Spec.Replicas,
			Status: response.DeploymentStatus{
				Replicas:          deployment.Status.Replicas,
				AvailableReplicas: deployment.Status.AvailableReplicas,
				ReadyReplicas:     deployment.Status.ReadyReplicas,
				UpdatedReplicas:   deployment.Status.UpdatedReplicas,
				Generation:        deployment.Status.ObservedGeneration,
			},
		}

		conditions := make([]response.DeploymentCondition, 0, len(deployment.Status.Conditions))

		for _, condition := range deployment.Status.Conditions {
			conditionData := response.DeploymentCondition{
				Name:               string(condition.Type),
				Status:             string(condition.Status),
				LastTransitionTime: utils.TrimKubeTime(condition.LastTransitionTime),
				Reason:             condition.Reason,
			}

			conditions = append(conditions, conditionData)
		}

		sort.Slice(conditions, func(i, j int) bool {
			return conditions[i].LastTransitionTime > conditions[j].LastTransitionTime
		})

		deploymentData.Conditions = conditions
		result = append(result, deploymentData)
	}

	return ctx.JSON(response.BaseResponse[response.Deployment]{
		TotalItems: len(result),
		Items:      result,
	})
}
