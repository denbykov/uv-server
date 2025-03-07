package business

import "fmt"

type Source int

const (
	Unknown Source = 0
	Youtube Source = 1
)

func (t Source) String() string {
	switch t {
	case Unknown:
		return "Unknown"
	case Youtube:
		return "Youtube"
	default:
		return fmt.Sprintf("Unknown: %d", t)
	}
}
