package config

import (
	"fmt"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"mighty/export"
)

type MightyConfig struct {
	MiteUrl           string `mapstructure:"url"`
	Token             string `mapstructure:"token"`
	EnableDebug       bool   `mapstructure:"debug"`
	EntriesHistory    string `mapstructure:"history"`
	CurrentExportFile *export.XlFile
	viper             *viper.Viper
}

const (
	DefaultCfgFile = ".mighty"
)

var (
	defaultCfgSearchPaths = []string{".", "$HOME"}

	CurrentConfig MightyConfig
	DefaultConfig = MightyConfig{
		MiteUrl:        "https://mite.yo.lk",
		Token:          "<get_your_token>",
		EnableDebug:    false,
		EntriesHistory: "4w",
	}
)

func SetupCfg(cfgFile string, generateCfg bool) {
	v := viper.New()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {

		v.SetEnvPrefix("MIGHTY")
		v.AutomaticEnv()

		v.SetConfigName(DefaultCfgFile)
		v.SetConfigType("yaml")

		for _, p := range defaultCfgSearchPaths {
			v.AddConfigPath(p)
		}
	}

	if generateCfg {
		v.SetDefault("url", DefaultConfig.MiteUrl)
		v.SetDefault("token", DefaultConfig.Token)
		v.SetDefault("debug", DefaultConfig.EnableDebug)
		v.SetDefault("history", DefaultConfig.EntriesHistory)

		err := v.SafeWriteConfig()
		if err != nil {
			logger.Fatalf("Unable to generate config %v", err)
			return
		}
		fmt.Printf("Generated config file at %s\n", v.ConfigFileUsed())
	} else {
		if err := v.ReadInConfig(); err != nil {
			logger.Fatalf("Unable to read the config, does config file exist? \nuse 'mighty gen --configFile' to create a config file\n%v", err)
		}

	}

	if err := v.Unmarshal(&CurrentConfig); err != nil {
		logger.Fatalf("Unable to read the config %v", err)
	}

	if CurrentConfig.EnableDebug {
		logger.SetLevel(logger.DebugLevel)
	} else {
		logger.SetLevel(logger.WarnLevel)
	}

	CurrentConfig.viper = v
	logger.Debugf("Using config file %s ", v.ConfigFileUsed())
	logger.Infof("Config: %v", v.AllSettings())

}
