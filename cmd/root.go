package cmd

import (
	"fmt"
	"github.com/go-errors/errors"
	"os"

	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "restake-authz-ledger",
	Short: "A small application to help manually make ledger authz work on REStake",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output, especially errors")

}

func RunE(f func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		err = f(cmd, args)
		if verbose && err != nil {
			stackedErr, ok := err.(*errors.Error)
			if ok {
				fmt.Println(stackedErr.ErrorStack())
				return err
			}
			fmt.Println(err)
		}

		return err
	}
}
