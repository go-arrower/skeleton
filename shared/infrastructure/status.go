package infrastructure

import (
	"context"
	"fmt"
	"time"
)

func getSystemStatus(di *Container, serverStartedAt time.Time) interface{} {
	uptime := time.Since(serverStartedAt).Round(time.Second)

	dbOnline := "online"
	_ = dbOnline
	err := di.PGx.Ping(context.Background())
	if err != nil {
		dbOnline = fmt.Errorf("err: %w", err).Error()
	}

	statusData := map[string]any{
		"status":           "online", // later: maintenance mode, degraded ect.
		"time":             time.Now(),
		"uptime":           fmt.Sprintf("%s", uptime),
		"gitCommit":        "", // todo
		"gitHash":          "", // todo
		"organisationName": di.Config.OrganisationName,
		"applicationName":  di.Config.ApplicationName,
		"instanceName":     di.Config.InstanceName,
		"debug":            di.Config.Debug,

		"web":      di.Config.Web,
		"database": dbStatus{Postgres: di.Config.Postgres, Status: dbOnline},
		// s3
		// REST API
		// feature flags
		// queues
		// memory consumption

		"failures": map[string]any{},
	}

	return statusData
}

type dbStatus struct {
	Postgres
	Status string `json:"status"`
	// average response time (?)
}
