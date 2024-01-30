package arcgis

import (
	"encoding/json"
	"time"
)

// Arcgis returns unix timestamp in milliseconds
type Date struct {
	*time.Time
}

func NewDateFromUnixMillis(unix int64) Date {
	t := time.Unix(0, unix*int64(time.Millisecond))
	return Date{&t}
}

func (d *Date) UnmarshalJSON(b []byte) error {
	var timestamp int64
	if err := json.Unmarshal(b, &timestamp); err != nil {
		return err
	}
	t := time.Unix(0, timestamp*int64(time.Millisecond))
	d.Time = &t

	return nil
}
