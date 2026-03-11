package app

import (
	"fmt"
	"os"
	"time"

	docs "connector-kube/docs"
	"connector-kube/internal/config"
	"connector-kube/internal/health"
	"connector-kube/internal/http/routes"
	"connector-kube/internal/kubernetes"
	"connector-kube/internal/swagger"
	t "connector-kube/internal/telemetry"

	"github.com/go-co-op/gocron"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

var scheduler = gocron.NewScheduler(time.UTC)
var healthChecker = health.HealthChecker{
	Scheduler: scheduler,
}

func LoadKubernetesClient() {
	kubernetes.Client = &kubernetes.KubernetesClient{
		InCluster:      config.Config.InCluster,
		KubeConfigPath: config.Config.KubeConfigPath,
	}
	err := kubernetes.Client.Init()

	if err != nil {
		t.Log.Error("Failed to create Kubernetes client", "error", err.Error())
		os.Exit(1)
	}
}

func LoadHealthChecker() {
	healthChecker.CreateSchedule()
	scheduler.StartAsync()
}

func LoadRoutes(app fiber.Router) {
	// /api routes
	routes.Health(app)
	routes.Ready(app)

	// /namespaces routes
	routes.ListNamespaces(app)

	// /pods routes
	routes.ListPods(app)

	// /workloads/* routes
	routes.ListDeployments(app)

	// Swagger routes
	swaggerHandler := LoadSwagger()
	app.Get("/swagger/*", swaggerHandler)
	app.Get("/docs", swagger.Handler("/swagger/doc.json"))
}

func LoadSwagger() fiber.Handler {
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", config.Config.ListenHost, config.Config.ListenPort)

	if config.Config.OtelServiceVersion != "" {
		docs.SwaggerInfo.Version = config.Config.OtelServiceVersion
	}

	return fiberSwagger.FiberWrapHandler(
		fiberSwagger.URL("/swagger/doc.json"),
		fiberSwagger.DocExpansion("list"),
		fiberSwagger.DeepLinking(true),
		fiberSwagger.PersistAuthorization(true),
	)
}
