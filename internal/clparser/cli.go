package clparser

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"p2p-snake/internal/log"
)

func Parse() string {
	pflag.StringP("config", "c", "config/config.json", "Config file path")
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Logger.Fatalf("Command line parser error: %v", err)
	}
	return viper.GetString("config")
}
