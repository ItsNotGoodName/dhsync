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

func (args SyncVideoInModeArgs) SunriseSunset(now time.Time) (time.Time, time.Time) {
	srise, sset := sunrise.SunriseSunset(args.Latitude, args.Longitude, now.Year(), now.Month(), now.Day())
	srise = srise.In(args.Location).Add(args.SunriseOffset)
	sset = sset.In(args.Location).Add(args.SunsetOffset)
	return srise, sset
}

// SyncVideoInMode uses basic day and night settings.
func SyncVideoInMode(ctx context.Context, c dahuarpc.Conn, args SyncVideoInModeArgs) (configmanager.ConfigArray[config.VideoInMode], error) {
	cfg, err := config.GetVideoInMode(ctx, c)
	if err != nil {
		return cfg, err
	}

	// Calculate sunrise and sunset
	sunrise, sunset := args.SunriseSunset(time.Now())
	timeSection := dahuarpc.NewTimeSection(1, dahuarpc.TimeSectionDuration(sunrise), dahuarpc.TimeSectionDuration(sunset))

	var changed bool

	// Apply to each channel
	for channelIdx := range cfg.Tables {
		// Sync SwitchMode
		if cfg.Tables[channelIdx].Data.SwitchMode() != config.SwitchModeSchedule {
			cfg.Tables[channelIdx].Data.SetSwitchMode(config.SwitchModeSchedule)
			changed = true
		}

		// Sync TimeSection
		if cfg.Tables[channelIdx].Data.TimeSection[0][0].String() != timeSection.String() {
			cfg.Tables[channelIdx].Data.TimeSection[0][0] = timeSection
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

// SyncVideoInMode2 uses the new time plan for setting day and night profiles. This only works on cameras that have the new interface.
func SyncVideoInMode2(ctx context.Context, c dahuarpc.Conn, args SyncVideoInModeArgs) (configmanager.ConfigArray[config.VideoInMode], error) {
	cfg, err := config.GetVideoInMode(ctx, c)
	if err != nil {
		return cfg, err
	}

	now := time.Now()

	var changed bool

	// Apply to each channel
	for channelIdx := range cfg.Tables {
		// Sync SwitchMode
		if cfg.Tables[channelIdx].Data.SwitchMode() != config.SwitchModeProfiles {
			cfg.Tables[channelIdx].Data.SetSwitchMode(config.SwitchModeProfiles)
			changed = true
		}

		for i := range 12 {
			monthIdx := (int(now.Month()-1) + i) % 12

			// The current month should get sunrise and sunset for the current day.
			// The following months should get it for the first day of that month.
			var d time.Time
			if i == 0 {
				d = now.AddDate(0, i, 0)
			} else {
				d = now.AddDate(0, i, -now.Day())
			}
			sunrise, sunset := args.SunriseSunset(d)

			// Calculate day and night profiles
			timeSections := []dahuarpc.TimeSection2{
				dahuarpc.NewTimeSection2(0, dahuarpc.TimeSectionDuration(sunrise), config.ProfileNight),
				dahuarpc.NewTimeSection2(dahuarpc.TimeSectionDuration(sunrise.Add(time.Second)), dahuarpc.TimeSectionDuration(sunset), config.ProfileDay),
				dahuarpc.NewTimeSection2(dahuarpc.TimeSectionDuration(sunset.Add(time.Second)), 86400000000000, config.ProfileNight),
			}

			// Sync day and night profiles
			if len(cfg.Tables[channelIdx].Data.TimeSectionV2[monthIdx]) != len(timeSections) {
				cfg.Tables[channelIdx].Data.TimeSectionV2[monthIdx] = timeSections
				changed = true
			} else {
				for _, ts := range timeSections {
					if !slices.ContainsFunc(cfg.Tables[channelIdx].Data.TimeSectionV2[monthIdx], func(e dahuarpc.TimeSection2) bool { return e.String() == ts.String() }) {
						cfg.Tables[channelIdx].Data.TimeSectionV2[monthIdx] = timeSections
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
