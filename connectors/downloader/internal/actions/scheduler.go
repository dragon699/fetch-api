package actions

import (
	"connector-downloader/internal/config"
	t "connector-downloader/internal/telemetry"
	"time"

	"github.com/go-co-op/gocron"
)

type ActionsRunner struct {
	Scheduler *gocron.Scheduler
}

func (instance *ActionsRunner) CreateSchedule() {
	t.Log.Info("Scheduling qBittorrent checks..")
	instance.run()

	if config.Config.TorrentActionsJobID == nil {
		if instance.Scheduler == nil {
			return
		}

		nextCheckTime := instance.getNextCheckTime()
		config.Config.TorrentActionsNextCheck = &nextCheckTime

		jobTag := "torrent_actions"
		job, _ := instance.Scheduler.Every(config.Config.TorrentActionsIntervalSeconds).Seconds().Do(instance.run)
		job.Tag(jobTag)
		config.Config.TorrentActionsJobID = &jobTag
	}
}

func (instance *ActionsRunner) getNextCheckTime() string {
	ts := time.Now().Add(
		time.Duration(
			config.Config.TorrentActionsIntervalSeconds,
		) * time.Second,
	)

	return ts.Format("2006-01-02T15:04:05")
}
