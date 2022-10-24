package cmd

import (
	logger "github.com/Hnampk/fabric-flogging"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var (
	RootCmd = &cobra.Command{
		Use:   "start",
		Short: "",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			cmdLogger.Infof("Hi there!")
		},
	}

	cmdLogger = logger.MustGetLogger("cmd")
)

func init() {
}
