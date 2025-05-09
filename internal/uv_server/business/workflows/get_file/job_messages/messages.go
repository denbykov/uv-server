package jobmessages

import "time"

type Request struct {
	Id *int64 `json:"id"`
}

type Result struct {
	Id        int       `json:"id"`
	Path      *string   `json:"path"`
	SourceUrl string    `json:"sourceUrl"`
	Source    string    `json:"source"`
	Status    string    `json:"status"`
	AddedAt   time.Time `json:"addedAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
