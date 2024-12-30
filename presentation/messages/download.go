package messages

import (
	"server/business/entities"
)

type DownloadMessage struct {
	headerSize int32
	header     *Header

	payload *entities.Download
}

func (m *DownloadMessage) HeaderSize() int32 {
	return m.headerSize
}

func (m *DownloadMessage) Header() *Header {
	return m.header
}

func (m *DownloadMessage) Payload() *entities.Download {
	return m.payload
}
