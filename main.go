package main

import (
	"context"
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

	c := dahuarpc.NewClient("192.168.6.14", os.Getenv("IPC_USERNAME"), os.Getenv("IPC_PASSWORD"))
	defer c.Close(context.Background())

	_, err = config.GetVideoInMode(ctx, c)
	if err != nil {
		log.Fatal("Failed to GetVideoInMode", err)
	}

	// fmt.Println("SN:", data)

	// fmt.Println("Build:", data.Build, "BuildDate:", data.BuildDate, "Version", data.Version)
	// for _, d := range data.Tables {
	// 	pp.Println(d)
	// }

	_, err = SyncVideoInMode2(ctx, c, SyncVideoInModeArgs{
		Location:      time.Local,
		Latitude:      48.751911,
		Longitude:     -122.478683,
		SunriseOffset: 0,
		SunsetOffset:  0,
	})
	if err != nil {
		log.Fatal("Failed to SyncVideoInMode: ", err)
	}
}
