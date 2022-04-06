package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print information about current delegations",
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

		if len(delegations) == 0 {
			fmt.Println("No delegations found")
		}
		for _, d := range delegations {
			asFloat, err := d.GetShares().Float64()
			if err != nil {
				return err
			}
			fmt.Printf("%s: %f\n", d.ValidatorAddress, asFloat/1000000.0)
		}

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
