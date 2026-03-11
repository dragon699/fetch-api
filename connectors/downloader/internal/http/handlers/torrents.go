package handlers

import (
	"errors"
	"slices"

	"connector-downloader/internal/config"
	"connector-downloader/internal/http/dto/request"
	"connector-downloader/internal/http/dto/response"
	"connector-downloader/internal/mapper"
	"connector-downloader/internal/qbittorrent"
	"connector-downloader/internal/tpb"

	"github.com/gofiber/fiber/v2"
)

func ListTorrents(ctx *fiber.Ctx) error {
	torrents, err := qbittorrent.Client.ListTorrents()

	if err != nil {
		var clientErr *config.ClientError
		if errors.As(err, &clientErr) {
			return ctx.Status(502).JSON(
				response.ErrorResponse{
					Error:            err.Error(),
					UpstreamResponse: clientErr.UpstreamResponse(),
				},
			)
		}

		return ctx.Status(500).JSON(
			response.ErrorResponse{
				Error: err.Error(),
			},
		)
	}

	result := mapper.TorrentsFromQBittorrent(torrents)

	return ctx.JSON(response.BaseResponse[response.Torrent]{
		TotalItems: len(result),
		Items:      result,
	})
}

func AddTorrent(ctx *fiber.Ctx) error {
	var reqPayload request.AddTorrentPayload

	if err := ctx.BodyParser(&reqPayload); err != nil {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "Invalid request payload",
			},
		)
	}

	if reqPayload.URL == "" {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "url is required",
			},
		)
	}

	manage := true
	findSubs := false
	notify := false

	if reqPayload.Category == "" {
		reqPayload.Category = "jellyfin"
	}

	if (reqPayload.Category == "jellyfin") && !(slices.Contains(reqPayload.Tags, "jellyfin:rename=pending")) {
		reqPayload.Tags = append(reqPayload.Tags, "jellyfin:rename=pending")
	}

	if len(reqPayload.Tags) == 0 {
		reqPayload.Tags = []string{}
	}

	if reqPayload.SavePath == "" {
		reqPayload.SavePath = config.Config.QBittorrentDefaultSavePath
	}

	if reqPayload.Manage != nil {
		manage = *reqPayload.Manage
	}

	if (manage) && !(slices.Contains(reqPayload.Tags, "fetch-api")) {
		reqPayload.Tags = append(reqPayload.Tags, "fetch-api")
	}

	if reqPayload.FindSubs != nil {
		findSubs = *reqPayload.FindSubs
	}

	if findSubs {
		if reqPayload.Category != "jellyfin" {
			return ctx.Status(400).JSON(
				response.ErrorResponse{
					Error: "find_subs can only be true when `category` is jellyfin",
				},
			)
		}

		reqPayload.Tags = append(reqPayload.Tags, "jellyfin:find_subs=pending")
	}

	if reqPayload.Notify != nil {
		notify = *reqPayload.Notify
	}

	if (notify) && !(slices.Contains(reqPayload.Tags, "slack:notify=pending")) {
		reqPayload.Tags = append(reqPayload.Tags, "slack:notify=pending")
	}

	err := qbittorrent.Client.AddTorrent(
		reqPayload.URL,
		reqPayload.Category,
		reqPayload.Tags,
		reqPayload.SavePath,
	)

	if err != nil {
		var clientErr *config.ClientError
		if errors.As(err, &clientErr) {
			return ctx.Status(502).JSON(
				response.ErrorResponse{
					Error:            err.Error(),
					UpstreamResponse: clientErr.UpstreamResponse(),
				},
			)
		}

		return ctx.Status(500).JSON(
			response.ErrorResponse{
				Error: err.Error(),
			},
		)
	}

	return ctx.JSON(
		response.SuccessResponse{
			Success: true,
			Message: "Request sent to qBittorrent!",
		},
	)
}

func AddTorrentTags(ctx *fiber.Ctx) error {
	var reqPayload request.AddTorrentTagsPayload

	if err := ctx.BodyParser(&reqPayload); err != nil {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "Invalid request payload",
			},
		)
	}

	if reqPayload.Hash == "" || len(reqPayload.Tags) == 0 {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "hash and at least one tag in tags are required",
			},
		)
	}

	err := qbittorrent.Client.AddTorrentTags(
		reqPayload.Hash,
		reqPayload.Tags,
	)

	if err != nil {
		var clientErr *config.ClientError
		if errors.As(err, &clientErr) {
			return ctx.Status(502).JSON(
				response.ErrorResponse{
					Error:            err.Error(),
					UpstreamResponse: clientErr.UpstreamResponse(),
				},
			)
		}

		return ctx.Status(500).JSON(
			response.ErrorResponse{
				Error: err.Error(),
			},
		)
	}

	return ctx.JSON(
		response.SuccessResponse{
			Success: true,
			Message: "Tag/s added to torrent successfully!",
		},
	)
}

func DeleteTorrentTags(ctx *fiber.Ctx) error {
	var reqPayload request.DeleteTorrentTagsPayload

	if err := ctx.BodyParser(&reqPayload); err != nil {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "Invalid request payload",
			},
		)
	}

	if reqPayload.Hash == "" || len(reqPayload.Tags) == 0 {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "hash and at least one tag in tags are required",
			},
		)
	}

	err := qbittorrent.Client.DeleteTorrentTags(
		reqPayload.Hash,
		reqPayload.Tags,
	)

	if err != nil {
		var clientErr *config.ClientError
		if errors.As(err, &clientErr) {
			return ctx.Status(502).JSON(
				response.ErrorResponse{
					Error:            err.Error(),
					UpstreamResponse: clientErr.UpstreamResponse(),
				},
			)
		}

		return ctx.Status(500).JSON(
			response.ErrorResponse{
				Error: err.Error(),
			},
		)
	}

	return ctx.JSON(
		response.SuccessResponse{
			Success: true,
			Message: "Tag/s deleted from torrent successfully!",
		},
	)
}

func SearchTorrents(ctx *fiber.Ctx) error {
	var searchParams request.SearchTorrentsParams

	if err := ctx.QueryParser(&searchParams); err != nil {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "Invalid query parameters",
			},
		)
	}

	if searchParams.Query == "" {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "query is required",
			},
		)
	}

	// 200 = All video media
	torrents, err := tpb.Client.SearchTorrents(200, searchParams.Query)

	if err != nil {
		var clientErr *config.ClientError
		if errors.As(err, &clientErr) {
			return ctx.Status(502).JSON(
				response.ErrorResponse{
					Error:            err.Error(),
					UpstreamResponse: clientErr.UpstreamResponse(),
				},
			)
		}

		return ctx.Status(500).JSON(
			response.ErrorResponse{
				Error: err.Error(),
			},
		)
	}

	result := mapper.TorrentsFromTPB(torrents)

	return ctx.JSON(response.BaseResponse[response.Torrent]{
		TotalItems: len(result),
		Items:      result,
	})
}
