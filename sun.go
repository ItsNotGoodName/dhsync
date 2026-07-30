package main

import (
	"context"
	"slices"
	"time"

	"github.com/nathan-osman/go-sunrise"

	"github.com/ItsNotGoodName/dhapi-go/dahuarpc"
	"github.com/ItsNotGoodName/dhapi-go/dahuarpc/modules/configmanager"
	"github.com/ItsNotGoodName/dhapi-go/dahuarpc/modules/configmanager/config"
)

type SyncVideoInModeArgs struct {
	Location      *time.Location
	Latitude      float64
	Longitude     float64
	SunriseOffset time.Duration
	SunsetOffset  time.Duration
}

func SyncVideoInMode(ctx context.Context, c dahuarpc.Conn, args SyncVideoInModeArgs) (configmanager.ConfigArray[config.VideoInMode], error) {
	cfg, err := config.GetVideoInMode(ctx, c)
	if err != nil {
		return cfg, err
	}

	// Calculate sunrise and sunset
	now := time.Now()
	sunrise, sunset := sunrise.SunriseSunset(args.Latitude, args.Longitude, now.Year(), now.Month(), now.Day())
	sunrise = sunrise.In(args.Location).Add(args.SunriseOffset)
	sunset = sunset.In(args.Location).Add(args.SunsetOffset)
	ts := dahuarpc.NewTimeSection(1, dahuarpc.TimeSectionDuration(sunrise), dahuarpc.TimeSectionDuration(sunset))

	var changed bool

	// Apply to each channel
	for i := range cfg.Tables {
		// Sync SwitchMode
		if cfg.Tables[i].Data.SwitchMode() != config.SwitchModeSchedule {
			cfg.Tables[i].Data.SetSwitchMode(config.SwitchModeSchedule)
			changed = true
		}

		// Sync TimeSection
		if cfg.Tables[i].Data.TimeSection[0][0].String() != ts.String() {
			cfg.Tables[i].Data.TimeSection[0][0] = ts
			changed = true
		}

	}

	if changed {
		if err := configmanager.SetConfigArray(ctx, c, cfg); err != nil {
			return cfg, err
		}
	}

	return cfg, nil
}

func SyncVideoInMode2(ctx context.Context, c dahuarpc.Conn, args SyncVideoInModeArgs) (configmanager.ConfigArray[config.VideoInMode], error) {
	cfg, err := config.GetVideoInMode(ctx, c)
	if err != nil {
		return cfg, err
	}

	// Calculate sunrise and sunset
	now := time.Now()
	sunrise, sunset := sunrise.SunriseSunset(args.Latitude, args.Longitude, now.Year(), now.Month(), now.Day())
	sunrise = sunrise.In(args.Location).Add(args.SunriseOffset)
	sunset = sunset.In(args.Location).Add(args.SunsetOffset)

	timeSections := []dahuarpc.TimeSection2{
		dahuarpc.NewTimeSection2(0, dahuarpc.TimeSectionDuration(sunrise), config.ProfileNight),
		dahuarpc.NewTimeSection2(dahuarpc.TimeSectionDuration(sunrise.Add(time.Second)), dahuarpc.TimeSectionDuration(sunset), config.ProfileDay),
		dahuarpc.NewTimeSection2(dahuarpc.TimeSectionDuration(sunset.Add(time.Second)), 86400000000000, config.ProfileNight),
	}

	var changed bool

	// Apply to each channel
	for tableIdx := range cfg.Tables {
		// Sync SwitchMode
		if cfg.Tables[tableIdx].Data.SwitchMode() != config.SwitchModeProfiles {
			cfg.Tables[tableIdx].Data.SetSwitchMode(config.SwitchModeProfiles)
			changed = true
		}

		for monthIdx := range 12 {
			// Sync day and night profiles
			if len(cfg.Tables[tableIdx].Data.TimeSectionV2[monthIdx]) != len(timeSections) {
				cfg.Tables[tableIdx].Data.TimeSectionV2[monthIdx] = timeSections
				changed = true
			} else {
				for _, ts := range timeSections {
					if !slices.ContainsFunc(cfg.Tables[tableIdx].Data.TimeSectionV2[monthIdx], func(e dahuarpc.TimeSection2) bool { return e.String() == ts.String() }) {
						cfg.Tables[tableIdx].Data.TimeSectionV2[monthIdx] = timeSections
						changed = true
						break
					}
				}
			}
		}
	}

	if changed {
		if err := configmanager.SetConfigArray(ctx, c, cfg); err != nil {
			return cfg, err
		}
	}

	return cfg, nil
}
