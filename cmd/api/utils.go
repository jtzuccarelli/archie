package main

import (
	"errors"
	"time"
)

func parseCallDates(from, to string) (*time.Time, *time.Time, error) {
	var start, end *time.Time
	if from != "" {
		parsed, err := time.Parse(time.DateOnly, from)
		if err != nil {
			return nil, nil, errors.New("date_from must be a valid date in YYYY-MM-DD format")
		}
		start = &parsed
	}
	if to != "" {
		parsed, err := time.Parse(time.DateOnly, to)
		if err != nil {
			return nil, nil, errors.New("date_to must be a valid date in YYYY-MM-DD format")
		}
		if start != nil && parsed.Before(*start) {
			return nil, nil, errors.New("date_to must be on or after date_from")
		}
		nextDay := parsed.AddDate(0, 0, 1)
		end = &nextDay
	}
	return start, end, nil
}
