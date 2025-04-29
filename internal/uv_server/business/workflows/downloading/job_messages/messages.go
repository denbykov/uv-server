package jobmessages

type Request struct {
	Url *string `json:"url"`
}

type Progress struct {
	Id         int64   `json:"id"`
	Percentage float64 `json:"percentage"`
}
