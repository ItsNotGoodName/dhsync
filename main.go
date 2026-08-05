package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ItsNotGoodName/dhapi-go/dahuarpc"
	"github.com/ItsNotGoodName/dhapi-go/dahuarpc/modules/configmanager/config"
	"github.com/goccy/go-yaml"
)

func main() {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		log.Fatalln("Failed to read config file:", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalln("Failed to unmarshal config:", err)
	}

	if err := cfg.Parse(); err != nil {
		log.Fatalln("Failed to parse config:", err)
	}

	ctx := context.Background()

	for _, camera := range cfg.Cameras {
		SyncCamera(ctx, camera)
	}
}

func SyncCamera(ctx context.Context, camera ConfigCamera) {
	c := dahuarpc.NewClient(camera.IP, camera.Username, camera.Password)
	defer c.Close(context.Background())

	data, err := config.GetVideoInMode(ctx, c)
	if err != nil {
		log.Println("ERROR: SyncVideoInMode:", err)
		return
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
		fmt.Println("Name:", camera.Name, "\nSwitchMode:", data.Tables[0].Data.SwitchMode(), "TimeSection:", data.Tables[0].Data.TimeSectionV2[0])

		data, err = SyncVideoInMode2(ctx, c, CreateDayNightTimeSection2(syncArgs))
		if err != nil {
			log.Println("ERROR: SyncVideoInMode2:", err)
			return
		}

		fmt.Println("SwitchMode:", data.Tables[0].Data.SwitchMode(), "TimeSection:", data.Tables[0].Data.TimeSectionV2[0])
	} else {
		fmt.Println("Name:", camera.Name, "\nSwitchMode:", data.Tables[0].Data.SwitchMode(), "TimeSection:", data.Tables[0].Data.TimeSection[0][0])

		data, err = SyncVideoInMode(ctx, c, CreateDayNightTimeSection(syncArgs))
		if err != nil {
			log.Println("Failed to SyncVideoInMode:", err)
			return
		}

		fmt.Println("SwitchMode:", data.Tables[0].Data.SwitchMode(), "TimeSection:", data.Tables[0].Data.TimeSection[0][0])
	}
}
