package downloading

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cjmessages "uv_server/internal/uv_server/business/common_job_messages"
	dmocks "uv_server/internal/uv_server/business/data/mocks"
	wfData "uv_server/internal/uv_server/business/workflows/downloading/data"
	bdmocks "uv_server/internal/uv_server/business/workflows/downloading/data/mocks"
	jobmessages "uv_server/internal/uv_server/business/workflows/downloading/job_messages"

	"uv_server/internal/uv_server/business/data"
)

type testGetSourceFromUrl_TableEntry struct {
	url    string
	source data.Source
	err    error
}

func newDownloadingWf() *DownloadingWf {
	wf := &DownloadingWf{}
	wf.log = logrus.New().WithField("layer", "Business")

	wf.injectInternalDependencies()

	return wf
}

func TestGetSourceFromUrl(t *testing.T) {
	wf := newDownloadingWf()

	testData := []testGetSourceFromUrl_TableEntry{
		{
			url:    "https://www.youtube.com/watch?v=2AB3_l0iqSk&list=PLk_klgt4LMVdcHAKqQ93bKtQ_r2YgsxlP&index=2&ab_channel=starsetonline",
			source: data.Youtube,
		},
		{
			url:    "https://www.youtube.com/watch?v=2AB3_l0iqSk",
			source: data.Youtube,
		},
		{
			url:    "https://youtu.be/2AB3_l0iqSk?si=IQwuKRVw5Ik569Ta",
			source: data.Youtube,
		},
		{
			url:    "https://you.be/2AB3_l0iqSk?si=IQwuKRVw5Ik569Ta",
			source: data.Unknown,
			err:    errors.New(""),
		},
		{
			url:    "",
			source: data.Unknown,
			err:    errors.New(""),
		},
	}

	for _, entry := range testData {
		source, err := wf.getSourceFromUrl(entry.url)

		if err != nil && entry.err == nil {
			t.Error(err)
		}

		if source != entry.source {
			t.Errorf("wrong source %v for url %v", source, entry.url)
		}
	}
}

type testNormalizeYoutubeUrl_TableEntry struct {
	url           string
	source        data.Source
	normalizedUrl string
	err           error
}

func TestNormalizeUrl(t *testing.T) {
	wf := newDownloadingWf()

	testData := []testNormalizeYoutubeUrl_TableEntry{
		{
			url:           "https://www.youtube.com/watch?v=2AB3_l0iqSk&list=PLk_klgt4LMVdcHAKqQ93bKtQ_r2YgsxlP&index=2&ab_channel=starsetonline",
			source:        data.Youtube,
			normalizedUrl: "https://www.youtube.com/watch?v=2AB3_l0iqSk",
		},
		{
			url:           "https://www.youtube.com/watch?v=2AB3_l0iqSk",
			source:        data.Youtube,
			normalizedUrl: "https://www.youtube.com/watch?v=2AB3_l0iqSk",
		},
		{
			url:           "https://youtu.be/2AB3_l0iqSk?si=IQwuKRVw5Ik569Ta",
			source:        data.Youtube,
			normalizedUrl: "https://www.youtube.com/watch?v=2AB3_l0iqSk",
		},
		{
			url:    "https://www.youtube.com/list=PLk_klgt4LMVdcHAKqQ93bKtQ_r2YgsxlP&index=2&ab_channel=starsetonline",
			source: data.Youtube,
			err:    errors.New(""),
		},
	}

	for _, entry := range testData {
		url, err := wf.normalizeUrl(entry.url, entry.source)

		if err != nil && entry.err == nil {
			t.Error(err)
		}

		if url != entry.normalizedUrl {
			t.Errorf("bad url normalization %v for url %v", url, entry.url)
		}
	}
}

func TestStartDownloading_AlreadyDownloaded(t *testing.T) {
	downloaderMock := new(StartDownloadingFromYoutubeMock)
	dbMock := dmocks.NewMockDatabase(t)

	wf := newDownloadingWf()
	wf.startDownloadingFromYoutube = func(
		downloaderWg *sync.WaitGroup,
		url string,
		storageDir string,
	) error {
		return downloaderMock.do(downloaderWg, url, storageDir)
	}
	wf.database = dbMock

	var downloaderWg sync.WaitGroup
	url := "https://www.youtube.com/watch?v=2AB3_l0iqSk"

	dbMock.On("GetFileByUrl", url).Return(&data.File{}, nil)

	err := wf.startDownloading(&downloaderWg, url)
	assert.NotNil(t, err, "operation should have failed")

	dbMock.AssertExpectations(t)
}

func TestStartDownloading_HappyPass(t *testing.T) {
	downloaderMock := new(StartDownloadingFromYoutubeMock)
	dbMock := dmocks.NewMockDatabase(t)

	wf := newDownloadingWf()
	wf.startDownloadingFromYoutube = func(
		downloaderWg *sync.WaitGroup,
		url string,
		storageDir string,
	) error {
		return downloaderMock.do(downloaderWg, url, storageDir)
	}
	wf.database = dbMock

	var downloaderWg sync.WaitGroup
	url := "https://www.youtube.com/watch?v=2AB3_l0iqSk"

	dbMock.On("GetFileByUrl", url).Return(nil, nil)
	storage := &data.Settings{StorageDir: "./storage"}
	dbMock.On("GetSettings").Return(storage, nil)
	downloaderMock.On("do", &downloaderWg, url, storage.StorageDir).Return(nil)

	fileId := int64(1)

	var file *data.File
	dbMock.On("InsertFile", mock.Anything).Return(fileId, nil).
		Run(func(args mock.Arguments) {
			file = args.Get(0).(*data.File)
		})

	err := wf.startDownloading(&downloaderWg, url)
	assert.Nil(t, err, "operation should not have failed")

	assert.Equal(t, file.SourceUrl, url)
	assert.Equal(t, file.Source, data.Youtube)
	assert.Equal(t, file.Status, data.FsDownloading)

	assert.Equal(t, wf.fileId, fileId)

	dbMock.AssertExpectations(t)
	downloaderMock.AssertExpectations(t)
}

func TestStartDownloadingFromYoutube(t *testing.T) {
	downloaderMock := bdmocks.NewMockDownloader(t)

	wf := newDownloadingWf()
	wf.downloader = downloaderMock

	var downloaderWg sync.WaitGroup
	url := "https://www.youtube.com/watch?v=2AB3_l0iqSk"
	const storage = "./storage"

	downloaderMock.On("Download", &downloaderWg, url, storage).Return(nil)

	err := wf.startDownloadingFromYoutube(&downloaderWg, url, storage)
	assert.Nil(t, err, "operation should not have failed")
	time.Sleep(time.Second)

	downloaderWg.Done()
	downloaderWg.Wait()

	downloaderMock.AssertExpectations(t)
}

func TestRun_StartDownloadingFailed(t *testing.T) {
	downloaderMock := new(StartDownloadingMock)

	jobIn := make(chan interface{}, 1)

	wf := newDownloadingWf()
	wf.jobIn = jobIn
	wf.startDownloading = func(
		downloaderWg *sync.WaitGroup,
		url string,
	) error {
		return downloaderMock.do(downloaderWg, url)
	}

	var wg sync.WaitGroup
	url := "https://www.youtube.com/watch?v=2AB3_l0iqSk"
	request := jobmessages.Request{Url: &url}

	error := "something gone wrong"

	downloaderMock.On("do", mock.Anything, url).Return(errors.New(error))

	wg.Add(1)
	go wf.Run(&wg, &request)
	wg.Wait()

	select {
	case msg := <-jobIn:
		tMsg := msg.(*cjmessages.Error)
		assert.Equal(t, tMsg.Reason, error)
	default:
		t.Error("missing expected message")
	}

	downloaderMock.AssertExpectations(t)
}

func TestRun_ContextCancelled(t *testing.T) {
	downloaderMock := new(StartDownloadingMock)
	dbMock := dmocks.NewMockDatabase(t)

	jobIn := make(chan interface{}, 1)
	downloaderOut := make(chan interface{}, 1)

	ctx, cancel := context.WithCancel(context.Background())

	wf := newDownloadingWf()
	wf.jobIn = jobIn
	wf.jobCtx = ctx
	wf.database = dbMock
	wf.downloaderOut = downloaderOut
	wf.fileId = 1
	wf.startDownloading = func(
		downloaderWg *sync.WaitGroup,
		url string,
	) error {
		return downloaderMock.do(downloaderWg, url)
	}

	var wg sync.WaitGroup
	url := "https://www.youtube.com/watch?v=2AB3_l0iqSk"
	request := jobmessages.Request{Url: &url}

	downloaderMock.On("do", mock.Anything, url).Return(nil)

	var file *data.File
	dbMock.On("DeleteFile", mock.Anything).Return(nil).
		Run(func(args mock.Arguments) {
			file = args.Get(0).(*data.File)
		})

	cancel()

	wg.Add(1)
	go wf.Run(&wg, &request)

	msg := <-jobIn
	tMsg := msg.(*jobmessages.Progress)
	assert.Equal(t, tMsg.Percentage, float64(0))

	wg.Wait()

	select {
	case msg := <-jobIn:
		_, ok := msg.(*cjmessages.Canceled)
		assert.Equal(t, ok, true)
	default:
		t.Error("missing expected message")
	}

	assert.Equal(t, file.Id, wf.fileId)

	downloaderMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
}

func TestRun_HappyPass(t *testing.T) {
	downloaderMock := new(StartDownloadingMock)
	dbMock := dmocks.NewMockDatabase(t)

	jobIn := make(chan interface{}, 1)
	downloaderOut := make(chan interface{}, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wf := newDownloadingWf()
	wf.jobIn = jobIn
	wf.jobCtx = ctx
	wf.database = dbMock
	wf.downloaderOut = downloaderOut
	wf.fileId = 1
	wf.startDownloading = func(
		downloaderWg *sync.WaitGroup,
		url string,
	) error {
		return downloaderMock.do(downloaderWg, url)
	}

	var wg sync.WaitGroup
	url := "https://www.youtube.com/watch?v=2AB3_l0iqSk"
	request := jobmessages.Request{Url: &url}

	downloaderMock.On("do", mock.Anything, url).Return(nil)

	var updateFilePathFile *data.File
	dbMock.On("UpdateFilePath", mock.Anything).Return(nil).
		Run(func(args mock.Arguments) {
			updateFilePathFile = args.Get(0).(*data.File)
		})

	var UpdateFileStatus *data.File
	dbMock.On("UpdateFileStatus", mock.Anything).Return(nil).
		Run(func(args mock.Arguments) {
			UpdateFileStatus = args.Get(0).(*data.File)
		})

	wg.Add(1)
	go wf.Run(&wg, &request)

	expectedPercentage := 33.0
	expectedPercentage2 := 33.0
	filename := "filename"

	go func() {
		time.Sleep(1200 * time.Millisecond)
		downloaderOut <- &wfData.Progress{Percentage: expectedPercentage}
		downloaderOut <- &wfData.Progress{Percentage: 50.0}
		time.Sleep(1200 * time.Millisecond)
		downloaderOut <- &wfData.Progress{Percentage: expectedPercentage2}
		downloaderOut <- &wfData.Done{Filename: filename}
	}()

	msg := <-jobIn
	tMsg := msg.(*jobmessages.Progress)
	assert.Equal(t, tMsg.Percentage, float64(0))

	msg = <-jobIn
	tMsg = msg.(*jobmessages.Progress)
	assert.Equal(t, tMsg.Percentage, expectedPercentage)

	msg = <-jobIn
	tMsg = msg.(*jobmessages.Progress)
	assert.Equal(t, tMsg.Percentage, expectedPercentage2)

	msg = <-jobIn
	tMsg = msg.(*jobmessages.Progress)
	assert.Equal(t, tMsg.Percentage, float64(100))

	msg = <-jobIn
	_, ok := msg.(*jobmessages.Done)
	assert.True(t, ok)

	wg.Wait()

	assert.Equal(t, updateFilePathFile.Id, wf.fileId)
	assert.Equal(t, updateFilePathFile.Path, sql.NullString{String: filename, Valid: true})
	assert.Equal(t, UpdateFileStatus.Id, wf.fileId)
	assert.Equal(t, UpdateFileStatus.Status, data.FsFinished)

	downloaderMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
}

func TestRun_DownloadingFailed(t *testing.T) {
	downloaderMock := new(StartDownloadingMock)
	dbMock := dmocks.NewMockDatabase(t)

	jobIn := make(chan interface{}, 1)
	downloaderOut := make(chan interface{}, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wf := newDownloadingWf()
	wf.jobIn = jobIn
	wf.jobCtx = ctx
	wf.database = dbMock
	wf.downloaderOut = downloaderOut
	wf.fileId = 1
	wf.startDownloading = func(
		downloaderWg *sync.WaitGroup,
		url string,
	) error {
		return downloaderMock.do(downloaderWg, url)
	}

	var wg sync.WaitGroup
	url := "https://www.youtube.com/watch?v=2AB3_l0iqSk"
	request := jobmessages.Request{Url: &url}

	downloaderMock.On("do", mock.Anything, url).Return(nil)

	var deleteFileFile *data.File
	dbMock.On("DeleteFile", mock.Anything).Return(nil).
		Run(func(args mock.Arguments) {
			deleteFileFile = args.Get(0).(*data.File)
		})

	wg.Add(1)
	go wf.Run(&wg, &request)

	go func() {
		downloaderOut <- &wfData.Error{Reason: "something went wrong"}
	}()

	msg := <-jobIn
	tMsg := msg.(*jobmessages.Progress)
	assert.Equal(t, tMsg.Percentage, float64(0))

	msg = <-jobIn
	teMsg := msg.(*cjmessages.Error)
	assert.Equal(t, teMsg.Reason, "something went wrong")

	wg.Wait()

	assert.Equal(t, deleteFileFile.Id, wf.fileId)

	downloaderMock.AssertExpectations(t)
	dbMock.AssertExpectations(t)
}
