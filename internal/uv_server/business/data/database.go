package data

import (
	gfw "uv_server/internal/uv_server/business/workflows/get_file/job_messages"
	gfsw "uv_server/internal/uv_server/business/workflows/get_files/job_messages"
)

type Database interface {
	GetFile(id int64) (*File, error)
	GetFileByUrl(url string) (*File, error)
	InsertFile(file *File) (int64, error)
	UpdateFileStatus(file *File) error
	UpdateFilePath(file *File) error
	DeleteFile(file *File) error
	DeleteFiles(ids []int64) error

	GetFilesForGFW(request *gfsw.Request) (*gfsw.Result, error)
	GetFileForGFW(request *gfw.Request) (*gfw.Result, error)
}
