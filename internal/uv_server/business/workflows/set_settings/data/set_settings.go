package data

import "sync"

type SetSettings interface {
	SetPathDir(wg *sync.WaitGroup, pathDir string)
}
