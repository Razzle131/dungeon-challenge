package config

import (
	"encoding/json"
	"os"
	"time"
)

type jsonConfig struct {
	Floors        int    `json:"Floors"`
	Monsters      int    `json:"Monsters"`
	OpenAt        string `json:"OpenAt"`
	DurationHours int    `json:"Duration"`
}

type Config struct {
	Floors       int
	Monsters     int
	OpenAt       int64
	DurationUnix int
}

const secondsPerHour = 3600

func MustLoad(configPath string) Config {
	file, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var jsonCfg jsonConfig
	err = json.NewDecoder(file).Decode(&jsonCfg)
	if err != nil {
		panic(err)
	}

	parsedTime, err := time.Parse("15:04:05", jsonCfg.OpenAt)
	if err != nil {
		panic(err)
	}

	if jsonCfg.Floors <= 0 {
		panic("number of floors cant be nonpositive")
	}

	return Config{
		Floors:       jsonCfg.Floors,
		Monsters:     jsonCfg.Monsters,
		OpenAt:       parsedTime.Unix(),
		DurationUnix: jsonCfg.DurationHours * secondsPerHour,
	}
}
