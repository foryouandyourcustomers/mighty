package config

import (
	"fmt"
	"github.com/leanovate/mite-go/domain"
	"github.com/mitchellh/go-homedir"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"mighty/api"
	"mighty/export"
)

type MightyConfig struct {
	MiteUrl        string `mapstructure:"url"`
	Token          string `mapstructure:"token"`
	EnableDebug    bool   `mapstructure:"debug"`
	EntriesHistory string `mapstructure:"history"`
	client         *api.Client
	exportFile     *export.XlFile
	viper          *viper.Viper
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

	CurrentConfig.client = createClientFromConfig()
	CurrentConfig.viper = v
	logger.Debugf("Using config file %s ", v.ConfigFileUsed())
	logger.Infof("Config: %v", v.AllSettings())

}

func createClientFromConfig() *api.Client {
	client, err := api.New(CurrentConfig.MiteUrl, CurrentConfig.Token)
	if err != nil {
		logger.Fatalf("Unable to create client %v", err)
	}
	return client
}

func (cfg *MightyConfig) SyncFile(excelFile string, onlyPull bool) error {
	excelFilePath, err := homedir.Expand(excelFile)
	if err != nil {
		return err
	}

	if !onlyPull {
		err = cfg.pushToFile(excelFilePath, domain.Today())
		if err != nil {
			return err
		}
	}

	err = cfg.pullToFile(excelFilePath)
	if err != nil {
		return err
	}

	return nil
}

func (cfg *MightyConfig) pushToFile(excelFilePath string, date domain.LocalDate) error {

	if cfg.exportFile == nil {
		cfg.exportFile = export.ExcelFile(excelFilePath)
	}

	err := cfg.exportFile.ReloadFromDisk()
	if err != nil {
		return err
	}

	entries := cfg.exportFile.ReadAllEntries(date)
	err = cfg.client.SendEntriesToMite(entries)
	if err != nil {
		return err
	}

	return nil
}

func (cfg *MightyConfig) pullToFile(excelFilePath string) error {
	if cfg.exportFile == nil {
		cfg.exportFile = export.ExcelFile(excelFilePath)
	}

	allHistoricEntries, err := cfg.client.FetchEntries(cfg.EntriesHistory)
	err = cfg.exportFile.SaveAllEntries(allHistoricEntries)
	if err != nil {
		return err
	}
	return nil
}
