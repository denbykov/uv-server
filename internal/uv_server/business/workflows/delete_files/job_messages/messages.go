package jobmessages

type Request struct {
	Ids []int64 `json:"ids"`
}

type Error struct {
	FailedIds []int64 `json:"failedIds"`
}
