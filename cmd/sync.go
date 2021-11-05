package cmd

import (
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"mighty/config"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
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

		initConfig()

		file, err := cmd.Flags().GetString("timesheet")
		if err != nil {
			logger.Fatal("Unable to read the file flag", err)
		}

		onlyPull, err := cmd.Flags().GetBool("onlyPull")
		if err != nil {
			return
		}

		err = config.CurrentConfig.SyncFile(file, onlyPull)
		if err != nil {
			logger.Fatalf("Unable to sync entries to file %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().Bool("onlyPull", false, "only pulls the data from mite, updated entries will be overwritten")
}
