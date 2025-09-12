package scanner

import (
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"time"

	"github.com/buckhx/tiles"
	"github.com/rs/zerolog"
)

// (lat_min, lon_min, lat_max, lon_max)
type BoundingBox struct {
	LatMin float64
	LonMin float64
	LatMax float64
	LonMax float64
}

type WplaceScannerSettings struct {
	TileServerUrl         string
	ZoomLevel             int
	OutputDirectory       string
	UserAgent             string
	MaxConcurrentRequests int64
	Frequency             time.Duration
	BBox                  BoundingBox
	HttpClient            *http.Client
	Log                   *zerolog.Logger
}

type WplaceScanner struct {
	settings   *WplaceScannerSettings
	httpClient *http.Client
	fetcher    *TileFetcher
	log        *zerolog.Logger

	tileRequestChannel chan *tiles.Tile
}

func NewWplaceScanner(settings *WplaceScannerSettings) *WplaceScanner {
	scanner := &WplaceScanner{
		settings:           settings,
		tileRequestChannel: make(chan *tiles.Tile),
		log:                settings.Log,
	}

	client := settings.HttpClient
	if client == nil {
		client = &http.Client{}
	}

	fetcher := NewTileFetcher(&TileFetcherSettings{
		Client:        client,
		UserAgent:     settings.UserAgent,
		ServerURL:     settings.TileServerUrl,
		MaxConcurrent: settings.MaxConcurrentRequests,
		Log:           settings.Log,
	})

	scanner.fetcher = fetcher

	return scanner
}

func (w *WplaceScanner) checkOutputDirectory() error {
	w.log.Info().
		Str("outputDirectory", w.settings.OutputDirectory).
		Msg("Making sure the output directory exists...")
	return os.MkdirAll(w.settings.OutputDirectory, 0755)
}

func (w *WplaceScanner) Run() error {
	w.log.Info().Msg("Starting.")
	w.log.Debug().Interface("config", w.settings).Msg("Our settings")

	if err := w.checkOutputDirectory(); err != nil {
		return err
	} 

	ticker := time.NewTicker(w.settings.Frequency)

	go w.fetcher.Start()

	// Do it the first time so we don't have to wait an entire hour.
	w.download()

	for range ticker.C {
		w.download()
	}

	return nil
}

func (w *WplaceScanner) download() {
	zoom := w.settings.ZoomLevel

	northwestTile := tiles.FromCoordinate(
		w.settings.BBox.LatMin,
		w.settings.BBox.LonMin,
		w.settings.ZoomLevel)
	southeastTile := tiles.FromCoordinate(
		w.settings.BBox.LatMax,
		w.settings.BBox.LonMax,
		w.settings.ZoomLevel)

	minX := min(northwestTile.X, southeastTile.X)
	maxX := max(northwestTile.X, southeastTile.X)
	minY := min(northwestTile.Y, southeastTile.Y)
	maxY := max(northwestTile.Y, southeastTile.Y)

	tileCount := (maxX - minX) * (maxY - minY)
	w.log.Info().Int("tileCount", tileCount).Msg("Scheduling batch of tile downloads")

	
	batchTime := time.Now().UTC()
	batchTimeString := batchTime.Format("2006-01-02 15-04-05Z")
	directory := fmt.Sprintf("%s/%s", w.settings.OutputDirectory, batchTimeString)

	tileMap := &TileMap{}

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			w.log.Debug().Int("x", x).Int("y", y).Msg("Scheduling tile download")
			tile := &tiles.Tile{X: x, Y: y, Z: zoom}
			receivedTile := w.getTile(tile)
			tileMap.Add(receivedTile)
			w.writeTile(receivedTile, directory)
		}
	}
}

func (w *WplaceScanner) emptyTile() *WplaceTile {
	empty1000 := image.NewRGBA(image.Rect(0,0,1000,1000))
	return &WplaceTile{
		Image: empty1000,
	}
}

func (w *WplaceScanner) writeTile(receivedTile *WplaceTile, directory string) {
	subdirectory := fmt.Sprintf("%s/%d", directory, receivedTile.Coords.X)
	err := os.MkdirAll(directory, 0755)
	if err != nil {
		w.log.Error().Err(err).Str("dir", directory).Msg("Failed to create directory for tiles")
		return
	}

	filename := fmt.Sprintf("%s/%d.png", subdirectory, receivedTile.Coords.Y)
	file, err := os.Create(filename)
	if err != nil {
		w.log.Error().Err(err).Str("file", filename).Msg("Failed to open file for writing")
		return
	}

	png.Encode(file, receivedTile.Image)
	w.log.Debug().Str("file", filename).Msg("Wrote file successfully")
}

func (w *WplaceScanner) getTile(tile *tiles.Tile) *WplaceTile {
	receivedTile, err := w.fetcher.FetchTile(tile)
	if err != nil {
		w.log.Error().Err(err).Msg("Using empty tile in place of 404 not found tile")
		receivedTile = w.emptyTile()
		receivedTile.Coords = tile
	}
	return receivedTile
}
