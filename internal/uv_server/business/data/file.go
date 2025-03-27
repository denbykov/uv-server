package data

import (
	"database/sql"
	"time"
)

type FileStatus string

const (
	FsPending     FileStatus = "p"
	FsDownloading FileStatus = "d"
	FsFinished    FileStatus = "f"
)

type File struct {
	Id        int64
	Path      sql.NullString
	SourceUrl string
	Source    Source
	Status    FileStatus
	AddedAt   time.Time
	UpdatedAt time.Time
}
