package jobmessages

type Request struct {
	Url *string `json:"url"`
}

type Progress struct {
	Percentage float64 `json:"percentage"`
}
