package handlers

import (
	"errors"

	"connector-downloader/internal/config"
	"connector-downloader/internal/http/dto/request"
	"connector-downloader/internal/http/dto/response"
	"connector-downloader/internal/mapper"
	"connector-downloader/internal/qbittorrent"
	"connector-downloader/internal/service"
	"connector-downloader/internal/torrent_indexer"

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

	if reqPayload.Manage != nil {
		manage = *reqPayload.Manage
	}

	if reqPayload.FindSubs != nil {
		findSubs = *reqPayload.FindSubs
	}

	if reqPayload.Notify != nil {
		notify = *reqPayload.Notify
	}

	err := service.AddTorrent(service.AddTorrentOptions{
		URL:      reqPayload.URL,
		Category: reqPayload.Category,
		Tags:     reqPayload.Tags,
		SavePath: reqPayload.SavePath,
		Manage:   manage,
		FindSubs: findSubs,
		Notify:   notify,
	})

	if errors.Is(err, service.ErrFindSubsNotJellyfin) {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: err.Error(),
			},
		)
	}

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
	torrents, err := torrent_indexer.Client.SearchTorrents(200, searchParams.Query)

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

	result := mapper.TorrentsFromTorrentIndexer(torrents)

	return ctx.JSON(response.BaseResponse[response.Torrent]{
		TotalItems: len(result),
		Items:      result,
	})
}
