package routes

import (
	_ "connector-slack/internal/http/dto/request"
	_ "connector-slack/internal/http/dto/response"
	"connector-slack/internal/http/handlers"

	"github.com/gofiber/fiber/v2"
)

const notificationsRouterName = "/notifications"

// SendNotification godoc
// @Summary      Send a custom notification to Slack
// @Description  Send a custom notification to Slack with specified options.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        request  body      request.NotificationPayload  true  "Custom notification payload"
// @Success      200      {object}  response.NotificationStatus
// @Failure      400      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Router       /notifications [post]
func SendNotification(router fiber.Router) {
	api := router.Group(notificationsRouterName)
	api.Post("/", handlers.SendNotification)
}

// SendGrafanaAlertNotification godoc
// @Summary      Forward a Grafana alert payload to send a Slack message
// @Description  Accepts a Grafana Alerting webhook payload and posts it to Slack as a message.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        request  body      request.GrafanaAlertNotificationPayload  true  "Webhook payload from Grafana Alerting"
// @Success      200      {object}  response.NotificationStatus
// @Failure      400      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Router       /notifications/grafana/alerts [post]
func SendGrafanaAlertNotification(router fiber.Router) {
	api := router.Group(notificationsRouterName)
	api.Post("/grafana/alerts", handlers.SendGrafanaAlertNotification)
}

// SendTorrentNotification godoc
// @Summary      Forward a connector-downloader notification payload to send a Slack message
// @Description  Accepts a payload from connector-downloader and posts it to Slack as a message.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        request  body      request.TorrentNotificationPayload  true  "Notification payload from connector-downloader"
// @Success      200      {object}  response.NotificationStatus
// @Failure      400      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Router       /notifications/connector-downloader/torrent [post]
func SendTorrentNotification(router fiber.Router) {
	api := router.Group(notificationsRouterName)
	api.Post("/connector-downloader/torrent", handlers.SendTorrentNotification)
}
