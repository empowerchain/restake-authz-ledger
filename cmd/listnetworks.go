package cmd

import (
	"fmt"
	"github.com/bjaanes/restake-authz-ledger/pkg/network"

	"github.com/spf13/cobra"
)

var listNetworksCmd = &cobra.Command{
	Use:   "list-networks",
	Short: "List supported networks",
	RunE: RunE(func(cmd *cobra.Command, args []string) error {
		fmt.Println("Fetching networks...")
		networks, err := network.GetNetworks()
		if err != nil {
			return err
		}

		fmt.Println("Below is a list of the supported networks")
		for _, n := range networks {
			fmt.Printf("%s (Identifier: %q)\n", n.Name, n.Identifier)
		}

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(listNetworksCmd)
}
