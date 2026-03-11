package app

import (
	"fmt"
	"time"

	docs "connector-downloader/docs"
	"connector-downloader/internal/actions"
	"connector-downloader/internal/config"
	"connector-downloader/internal/health"
	"connector-downloader/internal/http/routes"
	"connector-downloader/internal/jellyfin"
	"connector-downloader/internal/qbittorrent"
	"connector-downloader/internal/slack"
	"connector-downloader/internal/swagger"
	"connector-downloader/internal/tpb"

	"github.com/go-co-op/gocron"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

var scheduler = gocron.NewScheduler(time.UTC)
var healthChecker = health.HealthChecker{
	Scheduler: scheduler,
}
var ActionsRunner = actions.ActionsRunner{
	Scheduler: scheduler,
}

func LoadQBittorrentClient() {
	qbittorrent.Client = &qbittorrent.QBittorrentClient{}
	qbittorrent.Client.Init()
}

func LoadTPBClient() {
	tpb.Client = &tpb.TPBClient{}
	tpb.Client.Init()
}

func LoadSlackClient() {
	slack.Client = &slack.SlackClient{}
	slack.Client.Init()
}

func LoadJellyfinClient() {
	jellyfin.Client = &jellyfin.JellyfinClient{}
	jellyfin.Client.Init()
}

func LoadHealthChecker() {
	healthChecker.CreateSchedule()
	scheduler.StartAsync()
}

func LoadActionsRunner() {
	ActionsRunner.CreateSchedule()
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
	// /api routes
	routes.Health(app)
	routes.Ready(app)

	// /torrents routes
	routes.ListTorrents(app)
	routes.SearchTorrents(app)
	routes.AddTorrent(app)
	routes.AddTorrentTags(app)
	routes.DeleteTorrentTags(app)

	// Swagger routes
	swaggerHandler := LoadSwagger()
	app.Get("/swagger/*", swaggerHandler)
	app.Get("/docs", swagger.Handler("/swagger/doc.json"))
}
