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
	"github.com/ItsNotGoodName/dhsync"
	"github.com/Rican7/lieut"
	"github.com/k0kubun/pp"
)

var (
	configFile string
)

func main() {
	appInfo := lieut.AppInfo{
		Name:    "dhsync",
		Summary: "Sync day and night profile on Dahua cameras.",
		Version: dhsync.Version,
	}

	globalFlags := flag.NewFlagSet(appInfo.Name, flag.ExitOnError)
	globalFlags.StringVar(&configFile, "config", "config.yml", "Config file to load")

	app := lieut.NewMultiCommandApp(
		appInfo,
		globalFlags,
		os.Stdout,
		os.Stderr,
	)

	app.SetCommand(lieut.CommandInfo{Name: "sync"}, Sync, nil)
	app.SetCommand(lieut.CommandInfo{Name: "daemon"}, Daemon, nil)
	app.SetCommand(lieut.CommandInfo{Name: "verify"}, Verify, nil)

	exitCode := app.Run(context.Background(), os.Args[1:])

	os.Exit(exitCode)
}

func Sync(ctx context.Context, arguments []string) error {
	cfg, err := dhsync.ReadConfig(configFile)
	if err != nil {
		return err
	}

	for _, camera := range cfg.Cameras {
		err := SyncCamera(ctx, camera)
		if err != nil && errors.Is(err, context.Canceled) {
			return err
		}
	}

	return err
}

func Daemon(ctx context.Context, arguments []string) error {
	cfg, err := dhsync.ReadConfig(configFile)
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

		fmt.Println("Sleeping for 24 hours...")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(24 * time.Hour):
		}

		newCfg, err := dhsync.ReadConfig(configFile)
		if err != nil {
			fmt.Println("ERROR: reading config file:", err)
		} else {
			cfg = newCfg
		}
	}
}

func Verify(ctx context.Context, arguments []string) error {
	cfg, err := dhsync.ReadConfig(configFile)
	if err != nil {
		return err
	}

	for _, camera := range cfg.Cameras {
		err := PrintCamera(ctx, camera)
		if err != nil && errors.Is(err, context.Canceled) {
			return err
		}
	}

	return err
}

func PrintCamera(ctx context.Context, camera dhsync.ConfigCamera) error {
	c := dahuarpc.NewClient(camera.IP, camera.Username, camera.Password)
	defer c.Close(context.Background())

	data, err := config.GetVideoInMode(ctx, c)
	if err != nil {
		log.Println("ERROR: SyncVideoInMode:", err)
		return err
	}

	pp.Println(data.Tables)

	return nil
}

func SyncCamera(ctx context.Context, camera dhsync.ConfigCamera) error {
	c := dahuarpc.NewClient(camera.IP, camera.Username, camera.Password)
	defer c.Close(context.Background())

	data, err := config.GetVideoInMode(ctx, c)
	if err != nil {
		log.Println("ERROR: SyncVideoInMode:", err)
		return err
	}

	syncArgs := dhsync.SyncVideoInModeArgs{
		Timezone:      camera.Timezone_P,
		Latitude:      camera.Latitude,
		Longitude:     camera.Longitude,
		SunriseOffset: camera.Sunrise_Offset_P,
		SunsetOffset:  camera.Sunset_Offset_P,
	}

	// Check if time plan capable
	if len(data.Tables[0].Data.TimeSectionV2) == 12 {
		fmt.Println("SYNCING", camera.Name, "\n\tPREVIOUS SwitchMode:", data.Tables[0].Data.SwitchMode(), "TimeSection:", data.Tables[0].Data.TimeSectionV2[0])

		data, err = dhsync.SyncVideoInMode2(ctx, c, dhsync.CreateDayNightTimeSection2(syncArgs))
		if err != nil {
			log.Println("ERROR: SyncVideoInMode2:", err)
			return err
		}

		fmt.Println("\tCURRENT SwitchMode:", data.Tables[0].Data.SwitchMode(), "TimeSection:", data.Tables[0].Data.TimeSectionV2[0])
	} else {
		fmt.Println("SYNCING", camera.Name, "\n\tPREVIOUS SwitchMode:", data.Tables[0].Data.SwitchMode(), "TimeSection:", data.Tables[0].Data.TimeSection[0][0])

		data, err = dhsync.SyncVideoInMode(ctx, c, dhsync.CreateDayNightTimeSection(syncArgs))
		if err != nil {
			log.Println("Failed to SyncVideoInMode:", err)
			return err
		}

		fmt.Println("\tCURRENT SwitchMode:", data.Tables[0].Data.SwitchMode(), "TimeSection:", data.Tables[0].Data.TimeSection[0][0])
	}

	return nil
}
