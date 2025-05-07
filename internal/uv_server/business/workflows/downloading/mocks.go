package downloading

import (
	"sync"

	"github.com/stretchr/testify/mock"
)

type StartDownloadingFromYoutubeMock struct {
	mock.Mock
}

func (m *StartDownloadingFromYoutubeMock) do(
	downloaderWg *sync.WaitGroup,
	url string,
	storageDir string,
) error {
	args := m.Called(downloaderWg, url, storageDir)
	return args.Error(0)
}

type StartDownloadingMock struct {
	mock.Mock
}

func (m *StartDownloadingMock) do(
	downloaderWg *sync.WaitGroup,
	url string,
) error {
	args := m.Called(downloaderWg, url)
	return args.Error(0)
}
