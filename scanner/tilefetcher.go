package scanner

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/buckhx/tiles"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type TileFetcher struct {
	cfg                *viper.Viper
	settings           *TileFetcherSettings
	log                *zerolog.Logger
	tileRequestChannel chan *tiles.Tile
	tileReplies        map[string]chan *WplaceTile
	tileRepliesMu      sync.Mutex
}

type TileFetcherSettings struct {
	Client        *http.Client
	UserAgent     string
	ServerURL     string
	MaxConcurrent int64
	Log           *zerolog.Logger
}

type WplaceTile struct {
	Resp         *http.Response
	Coords       *tiles.Tile
	Image        image.Image
	RequestTime  time.Time
	ReceivedTime time.Time
	URL          *url.URL
	err          error
}

func NewTileFetcher(settings *TileFetcherSettings) *TileFetcher {
	return &TileFetcher{
		settings:           settings,
		tileRequestChannel: make(chan *tiles.Tile),
		log:                settings.Log,
		tileReplies:        make(map[string]chan *WplaceTile),
	}
}

func (t *TileFetcher) Start() {
	go t.tileRequestWorker()
}

func makeTileKey(tile *tiles.Tile) string {
	return fmt.Sprintf("%d-%d-%d", tile.X, tile.Y, tile.Z)
}

func (t *TileFetcher) MakeTileUrl(tile *tiles.Tile) string {
	base := t.settings.ServerURL
	withX := strings.ReplaceAll(base, "{x}", fmt.Sprint(tile.X))
	withY := strings.ReplaceAll(withX, "{y}", fmt.Sprint(tile.Y))
	withZ := strings.ReplaceAll(withY, "{z}", fmt.Sprint(tile.Z))
	return withZ
}

func (t *TileFetcher) FetchTile(tile *tiles.Tile) (*WplaceTile, error) {
	ch := make(chan *WplaceTile)
	key := makeTileKey(tile)

	t.tileRepliesMu.Lock()
	t.tileReplies[key] = ch
	t.tileRepliesMu.Unlock()

	t.tileRequestChannel <- tile

	receivedTile := <-ch

	return receivedTile, receivedTile.err
}

func (t *TileFetcher) tileRequestWorker() {
	var (
		maxWorkers = t.settings.MaxConcurrent
		sem        = semaphore.NewWeighted(int64(maxWorkers))
		ctx        = context.TODO()
	)

	for tile := range t.tileRequestChannel {
		if err := sem.Acquire(ctx, 1); err != nil {
			t.log.Err(err).Msg("Failed to acquire semaphore while queueing tile requests")
			break
		}

		go func(tile *tiles.Tile) {
			defer sem.Release(1)
			log := t.log.With().Str("url", t.MakeTileUrl(tile)).Logger()

			requestSentTime := time.Now().UTC()
			receivedTile, err := t.doTileRequest(tile)
			receivedTime := time.Now().UTC()
			if err != nil {
				log.Err(err).Msg("Failed to download tile")
				receivedTile.err = err
			}
			receivedTile.RequestTime = requestSentTime
			receivedTile.ReceivedTime = receivedTime

			t.tileRepliesMu.Lock()
			ch, ok := t.tileReplies[makeTileKey(tile)]
			if !ok {
				log.Warn().Msg("Failed to find reply channel for requested tile!")
			}
			ch <- receivedTile
			t.tileRepliesMu.Unlock()
		}(tile)
	}
}

func (t *TileFetcher) doTileRequest(tile *tiles.Tile) (*WplaceTile, error) {
	client := t.settings.Client
	req := http.Request{}
	req.Header = http.Header{}
	req.Header.Add("user-agent", t.settings.UserAgent)
	req.Method = http.MethodGet

	urlString := t.MakeTileUrl(tile)
	parsed, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	req.URL = parsed

	t.log.Info().Stringer("url", req.URL).Msg("Fetching tile URL")

	resp, err := client.Do(&req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	receivedTile := &WplaceTile{}
	receivedTile.URL = parsed
	receivedTile.Coords = tile
	receivedTile.Resp = resp

	if resp.StatusCode != http.StatusOK {
		return receivedTile, fmt.Errorf("failed to fetch tile: %s", resp.Status)
	}

	img, err := png.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image as png: %s", err)
	}
	receivedTile.Image = img

	t.log.Trace().Interface("respHeader", resp.Header).Msg("Fetched tile")

	return receivedTile, err
}
