package cmd

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"mighty/config"
	"os"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generates templates for quick start",
	Long: `Generates templates for quick start. 

Supports generating the template and has the following options (case sensitive):

* Configuration file:
$ mighty gen config  # uses the default file path
$ mighty gen config --configfile /path/to/file


* timesheet file:
$ mighty gen timesheet
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			err := cmd.Help()
			if err != nil {
				return err
			}
			os.Exit(0)
		} else if len(args) != 1 {
			return errors.New("Incorrect number of args.  Either use `config` or `timesheet` ")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		op := args[0]
		switch op {
		case "config":
			config.SetupCfg("", true)
		default:
			log.Errorf("Unsupported option %s", op)
			_ = cmd.Help()
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
}
