package main

import (
	"io"
	"os"
	"time"

	"github.com/a-random-lemurian/wplace-scanner/scanner"

	"github.com/carlmjohnson/versioninfo"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cmdRoot = &cobra.Command{
	Use:     "wplace-scanner",
	Version: versioninfo.Short(),
}

func init() {
	cmdRoot.AddCommand(cmdStart)

	crF := cmdRoot.PersistentFlags()
	crF.StringP("config", "c", "config.yml", "Path to configuration file")
	crF.Bool("debug", false, "Print additional debugging and trace information")
}

func main() {
	cmdRoot.Execute()
}

var cmdStart = &cobra.Command{
	RunE:  runStart,
	Use:   "start",
	Short: "Start the scanner",
}

func runStart(cmd *cobra.Command, args []string) error {
	config, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	minimumLevel := zerolog.InfoLevel
	if debug, err := cmd.Flags().GetBool("debug"); err == nil && debug {
		minimumLevel = zerolog.TraceLevel
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out: os.Stderr,
		TimeFormat: "2006-01-02 15:04:05.000000",
	}

	// TODO add lumberjack config
	multiOut := io.MultiWriter(consoleWriter)

	zerolog.TimeFieldFormat = time.RFC3339Nano
	preLog := zerolog.New(multiOut).With().Timestamp().Str("service", "wplace-scanner").Logger()
	log := preLog.Level(minimumLevel)

	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigFile(config)
	err = v.ReadInConfig()
	if err != nil {
		return err
	}

	log.Info().Msg("Loaded configuration file.")

	settings := &scanner.WplaceScannerSettings{}

	settings.Log = &log

	// Get slice of float64
	bboxRaw := v.Get("bbox").([]any)
	var bboxProcessed []float64
	for _, v := range bboxRaw {
		bboxProcessed = append(bboxProcessed, v.(float64))
	}
	bbox := scanner.BoundingBox{
		LatMin: bboxProcessed[0],
		LonMin: bboxProcessed[1],
		LatMax: bboxProcessed[2],
		LonMax: bboxProcessed[3],
	}

	settings.BBox = bbox
	settings.Frequency = v.GetDuration("frequency")
	settings.OutputDirectory = v.GetString("output")
	settings.TileServerUrl = v.GetString("tile_server")
	settings.UserAgent = v.GetString("user_agent")
	settings.MaxConcurrentRequests = v.GetInt64("max_concurrency")
	settings.ZoomLevel = v.GetInt("zoom_level")
	settings.GenerateStitches = v.GetBool("stitch_tiles")

	scanner := scanner.NewWplaceScanner(settings)
	
	return scanner.Run()
}
