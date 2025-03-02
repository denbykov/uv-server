package data

type Downloader interface {
	RegisterOnProgress(func(msg *ProgressMessage))

	Download(url string) (string, error)
}
