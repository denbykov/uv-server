package data

type Error struct {
	Reason string
}

type Progress struct {
	Percentage float64
}

type Done struct {
	Filename string
}
