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

type OscPrefixType string

const (
	OscConstantType OscPrefixType = "constant"
	OscSerialType   OscPrefixType = "serial"
)

type OscPrefixOption struct {
	Name         string        `json:"name"`
	Prefix       string        `json:"prefix"`
	AugmentIndex bool          `json:"augment_index"`
	ArgumentType OscPrefixType `json:"argument_type"` // constant, serial
	ArgumentBase int           `json:"argument_base"` // if constant the constant value, if serial the start value
}

type Config struct {
	BorderColor      string            `json:"border_color"`
	BorderWidth      int               `json:"border_width"`
	OscPrefixOptions []OscPrefixOption `json:"osc_prefix_options"`
}

var DefaultConfig = Config{
	BorderColor: "#FFFFFF",
	BorderWidth: 5,
	OscPrefixOptions: []OscPrefixOption{
		{Name: "Option 1", Prefix: "/streamdeck/option_1", ArgumentType: "serial", ArgumentBase: 1},
		{Name: "Option 2", Prefix: "/streamdeck/option_2", ArgumentType: "constant", ArgumentBase: 1},
		{Name: "Option 3", Prefix: "/streamdeck/option_3", ArgumentType: "serial", ArgumentBase: 1},
		{Name: "Custom", Prefix: "", ArgumentType: "constant", ArgumentBase: 1},
	},
}

func LoadConfig() (*Config, error) {
	configFile, err := os.Open("config.json")
	if err != nil {
		err = SaveConfig(&DefaultConfig)
		if err != nil {
			return nil, err
		}
		return &DefaultConfig, nil
	}
	defer configFile.Close()

	var cfg Config
	err = json.NewDecoder(configFile).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	configFile, err := os.Create("config.json")
	if err != nil {
		return err
	}
	defer configFile.Close()

	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}
