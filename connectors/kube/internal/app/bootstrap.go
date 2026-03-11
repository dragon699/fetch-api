// connector-kube bootstrap
//
// @title           Kube Connector
// @description     A connector that runs queries against Kubernetes cluster.
// @BasePath        /
// @produce         json
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"connector-kube/internal/config"
	t "connector-kube/internal/telemetry"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var app = fiber.New(fiber.Config{
	AppName: "connector-kube",
})

func Run() error {
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))

	app.Use(t.TracingMiddleware())

	LoadKubernetesClient()
	LoadHealthChecker()
	LoadRoutes(app)

	t.Log.Info("Starting Fiber..",
		"host", config.Config.ListenHost,
		"port", config.Config.ListenPort,
		"service", config.Config.Name,
	)

	go func() {
		if err := app.Listen(
			fmt.Sprintf("%s:%d", config.Config.ListenHost, config.Config.ListenPort),
		); err != nil {
			t.Log.Error("Failed to start Fiber", "error", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := app.Shutdown(); err != nil {
		t.Log.Error("Failed to shutdown server", "error", err.Error())
		return err
	}

	return nil
}
