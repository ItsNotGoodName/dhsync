package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ItsNotGoodName/dhapi-go/dahuarpc"
	"github.com/ItsNotGoodName/dhapi-go/dahuarpc/modules/configmanager/config"
	"github.com/Rican7/lieut"
	"github.com/goccy/go-yaml"
)

var version = "dev"

var (
	configFile string
	daemon     bool
)

func main() {
	appInfo := lieut.AppInfo{
		Name:    "dhsync",
		Summary: "Sync day and night profile on Dahua cameras.",
		Version: version,
	}

	globalFlags := flag.NewFlagSet(appInfo.Name, flag.ExitOnError)
	globalFlags.StringVar(&configFile, "config", "config.yml", "Config file to load")
	globalFlags.BoolVar(&daemon, "daemon", false, "Run in daemon mode")

	app := lieut.NewSingleCommandApp(
		appInfo,
		Run,
		globalFlags,
		os.Stdout,
		os.Stderr,
	)

	exitCode := app.Run(context.Background(), os.Args[1:])

	os.Exit(exitCode)
}

func Run(ctx context.Context, arguments []string) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}

	for {
		for _, camera := range cfg.Cameras {
			err := SyncCamera(ctx, camera)
			if err != nil && errors.Is(err, context.Canceled) {
				return err
			}
		}

		if !daemon {
			break
		}

		fmt.Println("Sleeping for 24 hours...")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(24 * time.Hour):
		}

		newCfg, err := ReadConfig()
		if err != nil {
			fmt.Println("ERROR: reading config file:", err)
		} else {
			cfg = newCfg
		}
	}

	return err
}

func ReadConfig() (Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	if err := cfg.Parse(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func SyncCamera(ctx context.Context, camera ConfigCamera) error {
	c := dahuarpc.NewClient(camera.IP, camera.Username, camera.Password)
	defer c.Close(context.Background())

	data, err := config.GetVideoInMode(ctx, c)
	if err != nil {
		log.Println("ERROR: SyncVideoInMode:", err)
		return err
	}

	syncArgs := SyncVideoInModeArgs{
		Timezone:      camera.Timezone_P,
		Latitude:      camera.Latitude,
		Longitude:     camera.Longitude,
		SunriseOffset: camera.Sunrise_Offset_P,
		SunsetOffset:  camera.Sunset_Offset_P,
	}

	// Check if time plan capable
	if len(data.Tables[0].Data.TimeSectionV2) == 12 {
		fmt.Println("SYNCING", camera.Name, "\n\tPREVIOUS SwitchMode:", data.Tables[0].Data.SwitchMode(), "TimeSection:", data.Tables[0].Data.TimeSectionV2[0])

		data, err = SyncVideoInMode2(ctx, c, CreateDayNightTimeSection2(syncArgs))
		if err != nil {
			log.Println("ERROR: SyncVideoInMode2:", err)
			return err
		}

		fmt.Println("\tCURRENT SwitchMode:", data.Tables[0].Data.SwitchMode(), "TimeSection:", data.Tables[0].Data.TimeSectionV2[0])
	} else {
		fmt.Println("SYNCING", camera.Name, "\n\tPREVIOUS SwitchMode:", data.Tables[0].Data.SwitchMode(), "TimeSection:", data.Tables[0].Data.TimeSection[0][0])

		data, err = SyncVideoInMode(ctx, c, CreateDayNightTimeSection(syncArgs))
		if err != nil {
			log.Println("Failed to SyncVideoInMode:", err)
			return err
		}

		fmt.Println("\tCURRENT SwitchMode:", data.Tables[0].Data.SwitchMode(), "TimeSection:", data.Tables[0].Data.TimeSection[0][0])
	}

	return nil
}
