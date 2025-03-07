package commonJobMessages

type Error struct {
	Reason string `json:"reason"`
}

type Canceled struct {
	Reason string `json:"reason"`
}

type Done struct {
}
