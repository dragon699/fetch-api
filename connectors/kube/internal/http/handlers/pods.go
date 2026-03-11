package handlers

import (
	"sort"
	"strings"

	"common/utils"
	"connector-kube/internal/dto/request"
	"connector-kube/internal/dto/response"
	"connector-kube/internal/kubernetes"

	"github.com/gofiber/fiber/v2"
	v1 "k8s.io/api/core/v1"
)

func getPodContainerStatusState(state v1.ContainerState) response.PodContainerStatusState {
	if state.Waiting != nil {
		return response.PodContainerStatusState{
			Reason: state.Waiting.Reason,
		}
	}
	if state.Running != nil {
		return response.PodContainerStatusState{
			StartedAt: utils.TrimKubeTime(state.Running.StartedAt),
		}
	}
	if state.Terminated != nil {
		return response.PodContainerStatusState{
			StartedAt:  utils.TrimKubeTime(state.Terminated.StartedAt),
			FinishedAt: utils.TrimKubeTime(state.Terminated.FinishedAt),
			Reason:     state.Terminated.Reason,
			ExitCode:   int(state.Terminated.ExitCode),
		}
	}
	return response.PodContainerStatusState{}
}

func getPodUnhealthyStatus(status v1.ContainerStatus) (response.PodStatus, bool) {
	if status.State.Waiting != nil && status.State.Waiting.Reason != "" {
		return response.PodStatus{
			Status: status.State.Waiting.Reason,
		}, true
	}

	if status.State.Terminated != nil {
		if status.State.Terminated.ExitCode == 0 {
			return response.PodStatus{}, false
		}

		if status.State.Terminated.Reason != "" {
			return response.PodStatus{
				Status: status.State.Terminated.Reason,
			}, true
		}

		return response.PodStatus{
			Status: "Crashed",
		}, true
	}

	if status.State.Running != nil && !status.Ready {
		return response.PodStatus{
			Status: "Pending",
		}, true
	}

	return response.PodStatus{}, false
}

func ListPods(ctx *fiber.Ctx) error {
	var params request.ListPodsParams

	if err := ctx.QueryParser(&params); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(
			response.ErrorResponse{
				Error: err.Error(),
			},
		)
	}

	params.Status = strings.ToLower(params.Status)
	params.StatusNot = strings.ToLower(params.StatusNot)

	pods, err := kubernetes.Client.ListPods()

	if err != nil {
		return ctx.Status(500).JSON(
			response.ErrorResponse{
				Error: err.Error(),
			},
		)
	}

	result := make([]response.Pod, 0, len(pods))

	for _, pod := range pods {
		if params.Namespace != "" && pod.ObjectMeta.Namespace != params.Namespace {
			continue
		}

		podData := response.Pod{
			Name:      pod.ObjectMeta.Name,
			Namespace: pod.ObjectMeta.Namespace,
			CreatedAt: utils.TrimKubeTime(pod.ObjectMeta.CreationTimestamp),
			Status:    string(pod.Status.Phase),
		}

		owners := make([]response.PodOwner, 0, len(pod.ObjectMeta.OwnerReferences))
		containers := make([]response.PodContainer, 0, len(pod.Spec.Containers))
		conditions := make([]response.PodCondition, 0, len(pod.Status.Conditions))
		containerStatuses := make(map[string]v1.ContainerStatus, len(pod.Status.ContainerStatuses))

		if pod.Status.StartTime != nil {
			podData.StartedAt = utils.TrimKubeTime(*pod.Status.StartTime)
		}

		for _, owner := range pod.ObjectMeta.OwnerReferences {
			owners = append(owners, response.PodOwner{
				Name: owner.Name,
				Kind: owner.Kind,
			})
		}

		for _, status := range pod.Status.ContainerStatuses {
			containerStatuses[status.Name] = status
		}

		for _, container := range pod.Spec.Containers {
			if params.ContainerNameContains != "" && !strings.Contains(container.Name, params.ContainerNameContains) {
				continue
			}

			containerData := response.PodContainer{
				Name:  container.Name,
				Image: container.Image,
			}

			if status, exists := containerStatuses[container.Name]; exists {
				containerStatus := response.PodContainerStatus{
					Started:      status.Started != nil && *status.Started,
					Ready:        status.Ready,
					RestartCount: int(status.RestartCount),
					State:        getPodContainerStatusState(status.State),
				}

				if lastState := getPodContainerStatusState(status.LastTerminationState); lastState != (response.PodContainerStatusState{}) {
					containerStatus.LastTerminationState = &lastState
				}

				containerData.Status = containerStatus
			}

			containers = append(containers, containerData)
		}

		for _, status := range pod.Status.ContainerStatuses {
			if healthStatus, hasIssue := getPodUnhealthyStatus(status); hasIssue {
				podData.Status = healthStatus.Status
				break
			}
		}

		for _, condition := range pod.Status.Conditions {
			conditionData := response.PodCondition{
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

		podData.Owners = owners
		podData.Containers = containers
		podData.Conditions = conditions

		if params.JobsOnly {
			podIsJob := false

			for _, owner := range podData.Owners {
				if strings.ToLower(owner.Kind) == "job" {
					podIsJob = true
					break
				}
			}

			if !podIsJob {
				continue
			}
		}

		podStatusLower := strings.ToLower(podData.Status)

		if params.Status != "" && podStatusLower != params.Status {
			continue
		}

		if params.StatusNot != "" && podStatusLower == params.StatusNot {
			continue
		}

		if params.ContainerNameContains != "" && len(containers) == 0 {
			continue
		}

		result = append(result, podData)
	}

	return ctx.JSON(response.BaseResponse[response.Pod]{
		TotalItems: len(result),
		Items:      result,
	})
}
