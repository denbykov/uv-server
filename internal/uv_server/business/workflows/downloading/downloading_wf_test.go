package downloading

import (
	"errors"
	"testing"
	"uv_server/internal/uv_server/business/data"
)

type testGetSourceFromUrl_TableEntry struct {
	url    string
	source data.Source
	err    error
}

func TestGetSourceFromUrl(t *testing.T) {
	wf := &DownloadingWf{}

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

func TestNormalizeYoutubeUrl(t *testing.T) {
	wf := &DownloadingWf{}

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

	// wf.getSourceFromUrl =
}
