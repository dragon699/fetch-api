package health

import (
	"time"

	"connector-slack/internal/config"
	"connector-slack/internal/slack"
	t "connector-slack/internal/telemetry"

	"github.com/go-co-op/gocron"
)

type HealthChecker struct {
	Scheduler *gocron.Scheduler
}

func (instance *HealthChecker) CreateSchedule() {
	t.Log.Info("Scheduling health checks..")
	instance.updateStatus()

	if config.Config.HealthJobID == nil {
		instance.updateSchedule()
	}
}

func (instance *HealthChecker) getNextInterval() int {
	if config.Config.Healthy != nil && *config.Config.Healthy {
		return config.Config.HealthCheckIntervalSeconds
	} else {
		return config.Config.HealthRetryIntervalSeconds
	}
}

func (instance *HealthChecker) getNextCheckTime() string {
	ts := time.Now().Add(
		time.Duration(
			instance.getNextInterval(),
		) * time.Second,
	)

	return ts.Format("2006-01-02T15:04:05")
}

func (instance *HealthChecker) updateStatus() {
	var wasHealthy bool
	var isHealthy bool

	wasHealthyPtr := config.Config.Healthy

	if wasHealthyPtr != nil {
		wasHealthy = *wasHealthyPtr
	}

	lastCheckTime := time.Now().Format("2006-01-02T15:04:05")
	config.Config.HealthLastCheck = &lastCheckTime
	err := slack.Client.Ping()

	if err != nil {
		config.Config.Healthy = nil

		if wasHealthyPtr != nil {
			instance.updateSchedule()
		} else {
			nextCheckTime := instance.getNextCheckTime()
			config.Config.HealthNextCheck = &nextCheckTime
		}

		t.Log.Error("Health check failed, Slack is unreachable", "error", err.Error())

		return
	}

	isHealthy = true
	config.Config.Healthy = &isHealthy

	if wasHealthyPtr != nil && wasHealthy {
		nextCheckTime := instance.getNextCheckTime()
		config.Config.HealthNextCheck = &nextCheckTime
	} else {
		t.Log.Info("Health check successful")
		instance.updateSchedule()
	}
}

func (instance *HealthChecker) updateSchedule() {
	if instance.Scheduler == nil {
		return
	}

	if config.Config.HealthJobID != nil {
		instance.Scheduler.RemoveByTag(*config.Config.HealthJobID)
	}

	interval := instance.getNextInterval()
	nextCheckTime := instance.getNextCheckTime()
	config.Config.HealthNextCheck = &nextCheckTime

	jobTag := "health_check_slack"
	job, _ := instance.Scheduler.Every(interval).Seconds().Do(instance.updateStatus)
	job.Tag(jobTag)
	config.Config.HealthJobID = &jobTag
}
