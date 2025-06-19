package db

import (
	"fmt"
	"time"
)

// parseTime converts a database value to time.Time
func parseTime(data interface{}) (time.Time, error) {
	switch v := data.(type) {
	case []byte:
		return time.Parse("2006-01-02 15:04:05", string(v))
	case string:
		return time.Parse("2006-01-02 15:04:05", v)
	case time.Time:
		return v, nil
	default:
		return time.Time{}, fmt.Errorf("unsupported time format: %T", data)
	}
}