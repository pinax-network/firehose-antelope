package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{Use: "init", Short: "Initializes streaming fast's local environment", RunE: dfuseInitE}


func init() {
	RootCmd.AddCommand(initCmd)
}

func dfuseInitE(cmd *cobra.Command, args []string) (err error) {
	cmd.SilenceUsage = true
	return fmt.Errorf("disabled dfuse init command")
}