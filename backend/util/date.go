package util

import (
	"time"

	"github.com/gofiber/fiber/v2/log"
)

func ConvertToDate(timeReq string) time.Time {
	date, err := time.Parse("2006-01-02", timeReq)
	if err != nil {
		log.Errorf("failed to convert %v, err: %v", timeReq, err.Error())
		return time.Now()
	}

	return date
}
