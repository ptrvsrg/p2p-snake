package config

import (
	"path/filepath"

	"github.com/spf13/viper"

	"p2p-snake/internal/log"
)

type P2PMulticastConfig struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

type P2PConfig struct {
	Delay     int                `mapstructure:"delay"`
	Multicast P2PMulticastConfig `mapstructure:"multicast"`
}

type APIConfig struct {
	Port int `mapstructure:"port"`
}

type AllConfig struct {
	P2P P2PConfig `mapstructure:"p2p"`
	API APIConfig `mapstructure:"api"`
}

var Config AllConfig

func LoadConfig(filePath string) {
	viper.SetConfigFile(filePath)
	viper.SetConfigType(filepath.Ext(filePath)[1:])

	err := viper.ReadInConfig()
	if err != nil {
		log.Logger.Fatalf("Config error: %v", err)
	}

	err = viper.Unmarshal(&Config)
	if err != nil {
		log.Logger.Fatalf("Config error: %v", err)
	}
}
