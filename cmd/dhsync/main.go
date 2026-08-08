package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ItsNotGoodName/dhapi-go/dahuarpc"
	"github.com/ItsNotGoodName/dhapi-go/dahuarpc/modules/configmanager/config"
	"github.com/ItsNotGoodName/dhsync"
	"github.com/Rican7/lieut"
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

	app.SetCommand(lieut.CommandInfo{Name: "sync", Summary: "Sync day and night profiles."}, Sync, nil)
	app.SetCommand(lieut.CommandInfo{Name: "daemon", Summary: "Sync every 24 hours."}, Daemon, nil)
	app.SetCommand(lieut.CommandInfo{Name: "read", Summary: "Read current day and night profiles."}, Read, nil)

	exitCode := app.Run(context.Background(), os.Args[1:])

	os.Exit(exitCode)
}

func Sync(ctx context.Context, arguments []string) error {
	cfg, err := dhsync.ReadConfig(configFile)
	if err != nil {
		return err
	}

	passed := true
	for _, camera := range cfg.Cameras {
		err := SyncCamera(ctx, camera)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			passed = false
		}
	}

	if passed {
		err := Ping(ctx, cfg.Healthcheck_Url)
		if err != nil && errors.Is(err, context.Canceled) {
			return err
		}
	}

	return nil
}

func Daemon(ctx context.Context, arguments []string) error {
	cfg, err := dhsync.ReadConfig(configFile)
	if err != nil {
		return err
	}

	for {
		passed := true
		for _, camera := range cfg.Cameras {
			err := SyncCamera(ctx, camera)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				passed = false
			}
		}

		if passed {
			err := Ping(ctx, cfg.Healthcheck_Url)
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

func Read(ctx context.Context, arguments []string) error {
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

	for i, channel := range data.Tables {
		if len(channel.Data.TimeSectionV2) == 12 {
			fmt.Println("Name:", camera.Name, "Channel:", i+1, "SwitchMode:", channel.Data.SwitchMode(), "TimeSection:", channel.Data.TimeSectionV2[time.Now().In(camera.Timezone_P).Month()-1])
		} else {
			fmt.Println("Name:", camera.Name, "Channel:", i+1, "SwitchMode:", channel.Data.SwitchMode(), "TimeSection:", channel.Data.TimeSection[0][0])
		}
	}

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

	if len(data.Tables) == 0 {
		err := errors.New("camera has no channels")
		log.Println("ERROR:", err)
		return err
	}

	// Check if time plan capable
	if len(data.Tables[0].Data.TimeSectionV2) == 12 {
		monthIdx := time.Now().In(camera.Timezone_P).Month() - 1

		fmt.Println("SYNCING", camera.Name)
		for i := range data.Tables {
			fmt.Println("\tPREVIOUS Channel:", i+1, "SwitchMode:", data.Tables[i].Data.SwitchMode(), "TimeSection:", data.Tables[i].Data.TimeSectionV2[monthIdx])
		}

		data, err = dhsync.SyncVideoInMode2(ctx, c, dhsync.CreateDayNightTimeSection2(syncArgs))
		if err != nil {
			log.Println("ERROR: SyncVideoInMode2:", err)
			return err
		}

		for i := range data.Tables {
			fmt.Println("\tCURRENT Channel:", i+1, "SwitchMode:", data.Tables[i].Data.SwitchMode(), "TimeSection:", data.Tables[i].Data.TimeSectionV2[monthIdx])
		}
	} else {
		fmt.Println("SYNCING", camera.Name)
		for i := range data.Tables {
			fmt.Println("\tPREVIOUS Channel:", i+1, "SwitchMode:", data.Tables[i].Data.SwitchMode(), "TimeSection:", data.Tables[i].Data.TimeSection[0][0])
		}

		data, err = dhsync.SyncVideoInMode(ctx, c, dhsync.CreateDayNightTimeSection(syncArgs))
		if err != nil {
			log.Println("ERROR: SyncVideoInMode:", err)
			return err
		}

		for i := range data.Tables {
			fmt.Println("\tCURRENT Channel:", i+1, "SwitchMode:", data.Tables[i].Data.SwitchMode(), "TimeSection:", data.Tables[i].Data.TimeSection[0][0])
		}
	}

	return nil
}

func Ping(ctx context.Context, url string) error {
	if url == "" {
		return nil
	}

	var client = &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return err
	}
	_, err = client.Do(req)
	if err != nil {
		fmt.Println("ERROR: pinging health check url:", err)
	}

	return nil
}
