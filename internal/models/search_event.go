package models

import "time"

type SearchEvent struct {
	Search    string    `json:"search"`
	User      string    `json:"user"`
	Timestamp time.Time `json:"timestamp"`
}
