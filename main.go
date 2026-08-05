package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ItsNotGoodName/dhapi-go/dahuarpc"
	"github.com/ItsNotGoodName/dhapi-go/dahuarpc/modules/configmanager/config"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()

	ips := []string{
		"192.168.60.11",
		"192.168.60.12",
		"192.168.60.13",
		"192.168.60.14",
		"192.168.60.15",
		"192.168.60.16",
		"192.168.60.17",
	}

	for _, ip := range ips {
		c := dahuarpc.NewClient(ip, os.Getenv("IPC_USERNAME"), os.Getenv("IPC_PASSWORD"))
		defer c.Close(context.Background())

		data, err := config.GetVideoInMode(ctx, c)
		if err != nil {
			log.Fatal("Failed to SyncVideoInMode: ", err)
		}

		fmt.Println("IPS", ip, ":", data.Tables[0].Data.SwitchMode(), data.Tables[0].Data.TimeSection[0][0])

		_, err = SyncVideoInMode(ctx, c, CreateDayNightTimeSection(SyncVideoInModeArgs{
			Location:      time.Local,
			Latitude:      48.751911,
			Longitude:     -122.478683,
			SunriseOffset: 0,
			SunsetOffset:  0,
		}))
		if err != nil {
			log.Fatal("Failed to SyncVideoInMode: ", err)
		}

		c.Close(ctx)
	}

}
