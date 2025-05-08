package data

type Filesystem interface {
	DeleteFile(path string) error
}
