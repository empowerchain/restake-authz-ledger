package cmd

import (
	"github.com/spf13/cobra"
)

var grantCmd = &cobra.Command{
	Use:   "grant",
	Short: "Grant the required access to a validator",

	RunE: RunE(func(cmd *cobra.Command, args []string) error {
		network, err := selectNetwork()
		if err != nil {
			return err
		}

		addr, err := askForAddress()
		if err != nil {
			return err
		}

		delegations, err := network.GetDelegations(cmd.Context(), addr)
		if err != nil {
			return err
		}

		validator, err := selectValidator(network, delegations...)
		if err != nil {
			return err
		}

		if err := network.GrantRestake(addr, validator); err != nil {
			return err
		}

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(grantCmd)
}
