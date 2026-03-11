package kubernetes

import (
	"context"

	t "connector-kube/internal/telemetry"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var Client *KubernetesClient

type KubernetesClient struct {
	InCluster      bool
	KubeConfigPath string
	Client         *kubernetes.Clientset
}

func (instance *KubernetesClient) Init() error {
	var kubeConfig *rest.Config
	var err error

	if instance.InCluster {
		kubeConfig, err = rest.InClusterConfig()
	} else {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", instance.KubeConfigPath)
	}

	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(kubeConfig)

	if err != nil {
		return err
	}

	instance.Client = client
	return nil
}

func (instance *KubernetesClient) Ping() (string, int, error) {
	_, err := instance.Client.Discovery().ServerVersion()

	if err != nil {
		return "not_ok", 0, err
	}

	return "ok", 200, nil
}

func (instance *KubernetesClient) ListPods() ([]v1.Pod, error) {
	pods, err := instance.Client.CoreV1().Pods("").List(
		context.TODO(), metav1.ListOptions{},
	)

	if err != nil {
		t.Log.Error("Failed to list pods", "error", err.Error())
		return nil, err
	}

	return pods.Items, nil
}

func (instance *KubernetesClient) ListDeployments() ([]appsv1.Deployment, error) {
	deployments, err := instance.Client.AppsV1().Deployments("").List(
		context.TODO(), metav1.ListOptions{},
	)

	if err != nil {
		t.Log.Error("Failed to list deployments", "error", err.Error())
	}

	return deployments.Items, nil
}

func (instance *KubernetesClient) ListNamespaces() ([]v1.Namespace, error) {
	namespaces, err := instance.Client.CoreV1().Namespaces().List(
		context.TODO(), metav1.ListOptions{},
	)

	if err != nil {
		t.Log.Error("Failed to list namespaces", "error", err.Error())
	}

	return namespaces.Items, nil
}
