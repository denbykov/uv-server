package data

import "sync"

type Downloader interface {
	Download(wg *sync.WaitGroup, url string)
}
