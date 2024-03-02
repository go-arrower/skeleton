package pages

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
)

func formatAsDateOrTimeToday(t time.Time) string {
	now := time.Now()
	isToday := t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()

	createdAt := t.Format("2006.01.02 15:04")
	if isToday {
		createdAt = t.Format("15:04")
	}

	return createdAt
}

func timeAgo(t time.Time) string {
	if t.IsZero() {
		return "unclear"
	}

	seconds := time.Since(t).Nanoseconds()

	switch seconds := time.Duration(seconds); {
	case seconds < time.Minute:
		return "now"
	case seconds < 90*time.Minute:
		minutes := int(math.Round(float64(seconds / time.Minute)))
		if minutes == 1 {
			return fmt.Sprintf("%d minute ago", minutes)
		}

		return fmt.Sprintf("%d minutes ago", minutes)
	case seconds < 24*time.Hour:
		hours := int(math.Round(float64(seconds / time.Hour)))
		if hours == 1 {
			return fmt.Sprintf("%d hour ago", hours)
		}

		return fmt.Sprintf("%d hours ago", hours)
	case seconds < time.Hour*24*365:
		days := int(math.Round(float64(seconds / (time.Hour * 24))))
		if days == 1 {
			return fmt.Sprintf("%d day ago", days)
		}

		return fmt.Sprintf("%d days ago", days)
	default:
		years := int(math.Round(float64(seconds / (time.Hour * 24 * 365))))
		if years == 1 {
			return fmt.Sprintf("%d year ago", years)
		}

		return fmt.Sprintf("%d years ago", years)
	}
}

func prettyJobPayloadAsFormattedJSON(p []byte) string {
	return prettyJSON(p)
}

func prettyJobPayloadDataAsFormattedJSON(payload application.JobPayload) string {
	b, _ := json.Marshal(payload.JobData)
	return prettyJSON(b)
}

func prettyJSON(str []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, str, "", "  "); err != nil {
		return ""
	}

	return prettyJSON.String()
}
