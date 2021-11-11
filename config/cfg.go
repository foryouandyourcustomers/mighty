package config

import (
	"fmt"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"mighty/export"
	"os"
)

type MightyConfig struct {
	MiteUrl           string `mapstructure:"url"`
	Token             string `mapstructure:"token"`
	EnableDebug       bool   `mapstructure:"debug"`
	EntriesHistory    string `mapstructure:"history"`
	CurrentExportFile *export.XlFile
}

const (
	DefaultCfgFile = ".mighty"
)

var (
	v                     *viper.Viper
	home                  string
	defaultCfgSearchPaths []string
	CurrentConfig         MightyConfig
	DefaultConfig         = MightyConfig{
		MiteUrl:        "https://mite.yo.lk",
		Token:          "<get_your_token>",
		EnableDebug:    false,
		EntriesHistory: "4w",
	}
)

func init() {
	v = viper.New()
	home, _ = os.UserHomeDir()
	cwd, _ := os.Getwd()
	defaultCfgSearchPaths = []string{home, cwd}

}

func SetupCfg(cfgFile string, generateCfg bool) {
	//TODO: this is messy improve this?

	v.SetConfigType("yaml")

	if generateCfg {

		if v.ConfigFileUsed() == "" {
			cfgFile = fmt.Sprintf("%s/%s.yml", home, DefaultCfgFile)
		} else {
			cfgFile = v.ConfigFileUsed()
		}

		v.SetDefault("url", DefaultConfig.MiteUrl)
		v.SetDefault("token", DefaultConfig.Token)
		v.SetDefault("debug", DefaultConfig.EnableDebug)
		v.SetDefault("history", DefaultConfig.EntriesHistory)

		if err := v.SafeWriteConfigAs(cfgFile); err != nil {
			switch err.(type) {
			case viper.ConfigFileAlreadyExistsError:
				logger.Fatalf("Config %s already exists, nope, I won't override it. use `--config /new/file.yml` to use a different file ", cfgFile)
			default:
				logger.Fatal(err)
			}
		}
		logger.Infof("Generated config file at %s\n", cfgFile)
	} else {
		if cfgFile != "" {
			v.SetConfigFile(cfgFile)
		} else {
			v.SetEnvPrefix("MIGHTY")
			v.AutomaticEnv()
		}
		for _, p := range defaultCfgSearchPaths {
			v.AddConfigPath(p)
		}
	}

}

func ReadCfg() {
	logger.Infof("Using config file %s\n", v.ConfigFileUsed())

	if err := v.ReadInConfig(); err != nil {
		logger.Fatalf(`Unable to read the config, does config file exist? %s 
If it doesn't exist, use 'mighty gen --config' to create a config file`, err)
	}

	if err := v.Unmarshal(&CurrentConfig); err != nil {
		logger.Fatalf("Unable to read the config %v", err)
	}

	if CurrentConfig.EnableDebug {
		logger.SetLevel(logger.DebugLevel)
	} else {
		logger.SetLevel(logger.InfoLevel)
	}

	logger.Debugf("Config: %v", v.AllSettings())
}
