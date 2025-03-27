package data

type Database interface {
	GetFileByUrl(url string) (*File, error)
	InsertFile(file *File) (int64, error)
	UpdateFileStatus(file *File) error
	UpdateFilePath(file *File) error
	DeleteFile(file *File) error
}
