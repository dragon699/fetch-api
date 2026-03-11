package app

import (
	"fmt"
	"time"

	docs "connector-slack/docs"
	"connector-slack/internal/config"
	"connector-slack/internal/health"
	"connector-slack/internal/http/routes"
	"connector-slack/internal/slack"
	"connector-slack/internal/slack/socket"
	"connector-slack/internal/swagger"

	"github.com/go-co-op/gocron"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

var scheduler = gocron.NewScheduler(time.UTC)
var healthChecker = health.HealthChecker{
	Scheduler: scheduler,
}

func LoadHealthChecker() {
	healthChecker.CreateSchedule()
	scheduler.StartAsync()
}

func LoadSlackClient() {
	slack.Client = &slack.SlackClient{}
	slack.Client.Init()
}

func LoadSlackSocketMode() error {
	socket.SocketMode = &socket.SlackSocketMode{}
	socket.SocketMode.Init()

	return socket.SocketMode.Start()
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

func LoadRoutes(app fiber.Router) {
	routes.Health(app)
	routes.Ready(app)
	routes.SendNotification(app)
	routes.SendGrafanaAlertNotification(app)
	routes.SendTorrentNotification(app)

	swaggerHandler := LoadSwagger()
	app.Get("/swagger/*", swaggerHandler)
	app.Get("/docs", swagger.Handler("/swagger/doc.json"))
}

func UnloadSlackSocketMode() {
	socket.SocketMode.Stop()
}
