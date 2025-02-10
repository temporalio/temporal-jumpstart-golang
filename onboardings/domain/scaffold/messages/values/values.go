package values

import "time"

type DateRange struct {
	Min time.Time `json:"min"`
	Max time.Time `json:"max"`
}
