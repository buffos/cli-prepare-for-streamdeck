package config

import (
	"encoding/json"
	"os"
)

type MediaType int

const (
	ImageType MediaType = iota
	VideoType
	AudioType
)

type OscPrefixOption struct {
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
}

type Config struct {
	BorderColor      string            `json:"border_color"`
	BorderWidth      int               `json:"border_width"`
	OscPrefixOptions []OscPrefixOption `json:"osc_prefix_options"`
}

var DefaultConfig = Config{
	BorderColor: "#FF0000",
	BorderWidth: 5,
	OscPrefixOptions: []OscPrefixOption{
		{Name: "Option 1", Prefix: "/streamdeck/option_1"},
		{Name: "Option 2", Prefix: "/streamdeck/option_2"},
		{Name: "Option 3", Prefix: "/streamdeck/option_3"},
		{Name: "Custom", Prefix: ""},
	},
}

func LoadConfig() (*Config, error) {
	configFile, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var cfg Config
	err = json.NewDecoder(configFile).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
