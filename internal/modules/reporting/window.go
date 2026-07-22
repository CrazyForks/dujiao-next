package reporting

import (
	"errors"
	"strings"
	"time"
)

const CustomMaxDays = 90

var ErrRangeInvalid = errors.New("reporting range invalid")

// Query describes the time range shared by operational reports.
type Query struct {
	Range        string
	From         *time.Time
	To           *time.Time
	Timezone     string
	ForceRefresh bool
}

// Window is the normalized half-open interval [StartAt, EndAt).
type Window struct {
	Range    string
	StartAt  time.Time
	EndAt    time.Time
	Timezone string
}

// Resolve normalizes a reporting query against the supplied clock value.
func Resolve(input Query, now time.Time) (Window, error) {
	rangeKey := strings.ToLower(strings.TrimSpace(input.Range))
	if rangeKey == "" {
		rangeKey = "7d"
	}

	timezone := strings.TrimSpace(input.Timezone)
	location := time.Local
	if timezone != "" {
		if parsed, err := time.LoadLocation(timezone); err == nil {
			location = parsed
		} else {
			timezone = ""
		}
	}
	if timezone == "" {
		timezone = location.String()
	}

	localNow := now.In(location)
	todayStart := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, location)
	window := Window{Range: rangeKey, Timezone: timezone}

	switch rangeKey {
	case "today":
		window.StartAt = todayStart
		window.EndAt = todayStart.AddDate(0, 0, 1)
	case "7d":
		window.StartAt = todayStart.AddDate(0, 0, -6)
		window.EndAt = todayStart.AddDate(0, 0, 1)
	case "30d":
		window.StartAt = todayStart.AddDate(0, 0, -29)
		window.EndAt = todayStart.AddDate(0, 0, 1)
	case "custom":
		if input.From == nil || input.To == nil {
			return Window{}, ErrRangeInvalid
		}
		startAt := input.From.In(location)
		endAt := input.To.In(location)
		if endAt.Before(startAt) || endAt.Sub(startAt) > time.Hour*24*CustomMaxDays {
			return Window{}, ErrRangeInvalid
		}
		window.StartAt = startAt
		window.EndAt = endAt.Add(time.Second)
	default:
		return Window{}, ErrRangeInvalid
	}

	if !window.EndAt.After(window.StartAt) {
		return Window{}, ErrRangeInvalid
	}
	return window, nil
}
