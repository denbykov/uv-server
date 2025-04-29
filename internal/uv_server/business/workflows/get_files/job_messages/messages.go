package jobmessages

import "time"

type Request struct {
	Limit  *int `json:"limit"`
	Offset *int `json:"offset"`
}

type File struct {
	Id      int       `json:"id"`
	Source  string    `json:"source"`
	Status  string    `json:"status"`
	AddedAt time.Time `json:"addedAt"`
}

type Result struct {
	Files []File `json:"files"`
	Total int    `json:"total"`
}
