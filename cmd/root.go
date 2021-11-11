package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"mighty/config"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mighty",
	Short: "A mighty tool for keeping your mite entries upto date.",
	Long: `A small and dandy tool for keeping syncing your mite entries upto date. 
For example:

Keep a local copy of your entries in an excel file and keep it synced with mite as follows:

$ mighty sync  // keep the entries upto-date at the default excel file
$ mighty sync  --OnlyPull  --timesheet mite-entries.xlsx// keep the entries upto-date
`,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "config file (default $HOME/.mighty.yaml)")
	rootCmd.PersistentFlags().String("timesheet", "", "the file which stores the timesheet entries (default $HOME/entries.xlsx")
	log.SetOutput(os.Stdout)
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	cfgFile, err := rootCmd.Flags().GetString("config")
	if err != nil {
		log.Fatal("Unable to read config file", err)
	}
	config.SetupCfg(cfgFile, false)
}
