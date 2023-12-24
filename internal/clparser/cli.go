package clparser

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"p2p-snake/internal/log"
)

const (
	configOptionDescription     = "Config file path"
	visibilityOptionDescription = "Node visibility (the ability of clients to find this node using the hub)"
)

func Parse() (string, bool) {
	pflag.StringP("config", "c", "config/config.json", configOptionDescription)
	pflag.BoolP("visible", "v", false, visibilityOptionDescription)

	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Logger.Fatalf("Command line parser error: %v", err)
	}

	return viper.GetString("config"), viper.GetBool("visible")
}
