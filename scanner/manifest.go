package scanner

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type GeneratorInfo struct {
	Version     string `json:"version"`
	ProgramName string `json:"program"`
}

type Manifest struct {
	Generator GeneratorInfo `json:"generator"`
	Timestamp time.Time     `json:"timestamp"`
	TileCount int           `json:"tileCount"`
	Tiles     []TileInfo    `json:"tiles"`

	mu sync.Mutex
}

type TileInfo struct {
	Url          string    `json:"url"`
	Filename     string    `json:"filename"`
	LastModified time.Time `json:"lastModified"`
	RequestTime  time.Time `json:"requestTime"`
	ReceivedTime time.Time `json:"receivedTime"`
	HttpResponseCode int   `json:"httpResponseCode"`
}

func (m *Manifest) AddTile(tile *WplaceTile) {
	m.mu.Lock()
	defer m.mu.Unlock()

	urlString := ""
	if tile.URL != nil {
		urlString = tile.URL.String()
	}

	tileinfo := TileInfo{
		Url:      urlString,
		Filename: fmt.Sprintf("%d/%d.png", tile.Coords.X, tile.Coords.Y),
	}

	var lastModifiedTime time.Time = time.UnixMicro(0)

	lm := ""

	if tile.Resp != nil {
		lm = tile.Resp.Header.Get("last-modified")
		tileinfo.HttpResponseCode = tile.Resp.StatusCode
	}

	if lm != "" {
		t, err := http.ParseTime(lm)
		if err == nil {
			lastModifiedTime = t
		}
	}
	tileinfo.LastModified = lastModifiedTime

	tileinfo.RequestTime = tile.RequestTime
	tileinfo.ReceivedTime = tile.ReceivedTime


	m.Tiles = append(m.Tiles, tileinfo)
}
