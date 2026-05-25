package models

import "time"

type TopSearch struct {
	Search       string    `json:"search"`
	Count        int64     `json:"count"`
	LastSearched time.Time `json:"last_searched"`
}

func CmpTopSearches(a, b TopSearch) int {
	if a.Count > b.Count {
		return -1
	} else if a.Count < b.Count {
		return 1
	} else if a.LastSearched.After(b.LastSearched) {
		return -1
	} else if b.LastSearched.After(a.LastSearched) {
		return 1
	} else {
		return 0
	}
}

func NewTopSearch(search string, count int64, lastSearched time.Time) TopSearch {
	return TopSearch{
		Search:       search,
		Count:        count,
		LastSearched: lastSearched,
	}
}
