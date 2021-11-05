package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"mighty/config"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generates required documents",
	PreRun: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Generates the configuration or timesheet entries\n")
	},
	Run: func(cmd *cobra.Command, _ []string) {

		generateConfigFile, err := cmd.Flags().GetBool("configFile")
		if err != nil {
			log.Fatal("Unable to read config file")
		}

		if generateConfigFile {
			config.SetupCfg("", true)
		}

	},
}

func init() {
	rootCmd.AddCommand(genCmd)
	genCmd.Flags().Bool("configFile", false, "generates the empty config file (default is $HOME/.mighty.yaml)")
	genCmd.Flags().Bool("timesheetFile", false, "generates the excel sheet containing the timesheet entries")
}
