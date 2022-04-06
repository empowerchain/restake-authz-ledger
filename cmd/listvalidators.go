package cmd

import (
	"fmt"
	"github.com/bjaanes/restake-authz-ledger/pkg/validator"
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

var validatorsNetwork string

var listValidatorsCmd = &cobra.Command{
	Use:   "list-validators",
	Short: "List validators",
	RunE: RunE(func(cmd *cobra.Command, args []string) error {
		if validatorsNetwork == "" {
			return errors.New("--network required")
		}

		vs, err := validator.GetSupportedValidators(validatorsNetwork)
		if err != nil {
			return err
		}

		if len(vs) == 0 {
			fmt.Printf("No validators found for network %q (have you checked if the network identifier is correct?\n", validatorsNetwork)
		}

		for _, v := range vs {
			fmt.Printf("%s: %s\n", v.Path, v.Name)
		}

		return nil
	}),
}

func init() {
	listValidatorsCmd.Flags().StringVar(&validatorsNetwork, "network", "", "The network to list validators for")
	rootCmd.AddCommand(listValidatorsCmd)
}
