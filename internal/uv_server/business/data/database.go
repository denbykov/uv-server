package data

import (
	gfw "uv_server/internal/uv_server/business/workflows/get_files/job_messages"
)

type Database interface {
	GetFileByUrl(url string) (*File, error)
	InsertFile(file *File) (int64, error)
	UpdateFileStatus(file *File) error
	UpdateFilePath(file *File) error
	DeleteFile(file *File) error

	GetFilesForGFW(request *gfw.Request) (*gfw.Result, error)
}
