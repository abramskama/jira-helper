package jira

import (
	"errors"
	"time"
)

var (
	ErrParseDatFormat = errors.New("cant't parse date, the layout must be YYYY-MM-DD|today|yest")
)

func convertDate(date string) (string, error) {
	if date == "today" {
		return time.Now().Format(time.DateOnly), nil
	} else if date == "yest" {
		return time.Now().AddDate(0, 0, -1).Format(time.DateOnly), nil
	}

	if _, err := time.Parse(time.DateOnly, date); err != nil {
		return "", ErrParseDatFormat
	}

	return date, nil
}
