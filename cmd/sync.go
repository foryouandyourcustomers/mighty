package cmd

import (
	"github.com/leanovate/mite-go/domain"
	"github.com/mitchellh/go-homedir"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"mighty/api"
	"mighty/config"
	"mighty/export"
)

// syncCmd represents the sync command
var (
	client        *api.Client
	currentConfig config.MightyConfig

	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Syncs the timesheet entries from an excel sheet file defined by --timesheet parameter",
		Long: `Syncs the timesheet entries from an excel sheet file defined by --timesheet parameter

Note that the syncing only works for current month. Use '--onlyPull' to fetch the past entries for the correct format.
A typical workflow can be: 

$ mighty sync --onlyPull mite-entries.xlsx
$ // update the entries 
$ mighty sync mite-entries.xlsx

`,
		Run: func(cmd *cobra.Command, _ []string) {

			config.ReadCfg()

			file, err := cmd.Flags().GetString("timesheet")
			if err != nil {
				logger.Fatal("Unable to read the file flag", err)
			}

			onlyPull, err := cmd.Flags().GetBool("onlyPull")
			if err != nil {
				return
			}
			client, err = createClientFromConfig()
			if err != nil {
				logger.Fatalf("Unable to create api client %v", err)
			}

			currentConfig = config.CurrentConfig

			err = syncFile(file, onlyPull)
			if err != nil {
				logger.Fatalf("Unable to sync entries to file %v", err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().Bool("onlyPull", false, "only pulls the data from mite, updated entries will be overwritten")
}

func createClientFromConfig() (*api.Client, error) {
	client, err := api.New(config.CurrentConfig.MiteUrl, config.CurrentConfig.Token)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func syncFile(excelFile string, onlyPull bool) error {
	excelFilePath, err := homedir.Expand(excelFile)
	if err != nil {
		return err
	}

	if !onlyPull {
		err = pushToFile(excelFilePath, domain.Today())
		if err != nil {
			return err
		}
	}

	err = pullToFile(excelFilePath)
	if err != nil {
		return err
	}

	return nil
}

func pushToFile(excelFilePath string, date domain.LocalDate) error {
	exportFile := currentConfig.CurrentExportFile

	if exportFile == nil {
		exportFile = export.ExcelFile(excelFilePath)
	}

	err := exportFile.ReloadFromDisk()
	if err != nil {
		return err
	}

	entries := exportFile.ReadAllEntries(date)
	err = client.SendEntriesToMite(entries)
	if err != nil {
		return err
	}

	return nil
}

func pullToFile(excelFilePath string) error {
	exportFile := currentConfig.CurrentExportFile

	if exportFile == nil {
		exportFile = export.ExcelFile(excelFilePath)
	}

	allHistoricEntries, err := client.FetchEntries(currentConfig.EntriesHistory)
	err = exportFile.SaveAllEntries(allHistoricEntries)
	if err != nil {
		return err
	}

	pMap, sMap, err := client.FetchServiceProjects()
	if err != nil {
		return err
	}

	exportFile.SaveServiceProjects(pMap, sMap)
	return nil
}
