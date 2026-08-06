package dhsync

import "time"

type Config struct {
	Username string
	Password string

	Latitude  float64
	Longitude float64
	Timezone  string

	Sunrise_Offset string
	Sunset_Offset  string

	Cameras []ConfigCamera
}

type ConfigCamera struct {
	Name             string
	IP               string
	Username         string
	Password         string
	Latitude         float64
	Longitude        float64
	Timezone         string
	Timezone_P       *time.Location `yaml:"-"`
	Sunrise_Offset   string
	Sunrise_Offset_P time.Duration `yaml:"-"`
	Sunset_Offset    string
	Sunset_Offset_P  time.Duration `yaml:"-"`
}

func (c *Config) Parse() error {
	for i := range c.Cameras {
		// Default
		if c.Cameras[i].Name == "" {
			c.Cameras[i].Name = c.Cameras[i].IP
		}
		if c.Cameras[i].Username == "" {
			c.Cameras[i].Username = c.Username
		}
		if c.Cameras[i].Password == "" {
			c.Cameras[i].Password = c.Password
		}
		if c.Cameras[i].Longitude == 0 {
			c.Cameras[i].Longitude = c.Longitude
		}
		if c.Cameras[i].Latitude == 0 {
			c.Cameras[i].Latitude = c.Latitude
		}
		if c.Cameras[i].Timezone == "" {
			c.Cameras[i].Timezone = c.Timezone
		}
		if c.Cameras[i].Sunset_Offset == "" {
			c.Cameras[i].Sunset_Offset = c.Sunset_Offset
		}
		if c.Cameras[i].Sunrise_Offset == "" {
			c.Cameras[i].Sunrise_Offset = c.Sunrise_Offset
		}

		// Parse
		if c.Cameras[i].Timezone == "" {
			c.Cameras[i].Timezone_P = time.Local
		} else {
			tz, err := time.LoadLocation(c.Cameras[i].Timezone)
			if err != nil {
				return err
			}
			c.Cameras[i].Timezone_P = tz
		}

		if c.Cameras[i].Sunrise_Offset != "" {
			d, err := time.ParseDuration(c.Cameras[i].Sunrise_Offset)
			if err != nil {
				return err
			}
			c.Cameras[i].Sunrise_Offset_P = d
		}

		if c.Cameras[i].Sunset_Offset != "" {
			d, err := time.ParseDuration(c.Cameras[i].Sunset_Offset)
			if err != nil {
				return err
			}
			c.Cameras[i].Sunset_Offset_P = d
		}
	}

	return nil
}
