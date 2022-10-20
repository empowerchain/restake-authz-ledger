package cmd

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/bjaanes/restake-authz-ledger/pkg/network"
	"github.com/bjaanes/restake-authz-ledger/pkg/validator"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/go-errors/errors"
)

func askForAddress() (string, error) {
	addr := ""
	prompt := &survey.Input{
		Message: "Address",
	}
	if err := survey.AskOne(prompt, &addr); err != nil {
		return "", err
	}

	return addr, nil
}

func selectNetwork() (network.Network, error) {
	fmt.Println("Fetching networks...")
	networks, err := network.GetNetworks()
	if err != nil {
		return network.Network{}, errors.Wrap(err, 0)
	}

	networkNames := make([]string, len(networks))
	for i, n := range networks {
		networkNames[i] = n.Name
	}
	var networkChoiceIndex int
	networkPrompt := &survey.Select{
		Message: "Choose network",
		Options: networkNames,
	}
	if err := survey.AskOne(networkPrompt, &networkChoiceIndex); err != nil {
		return network.Network{}, errors.Wrap(err, 0)
	}

	return networks[networkChoiceIndex], nil
}

func selectValidator(n network.Network, delegations ...stakingtypes.Delegation) (validator.ValidatorForNetwork, error) {
	supportedValidators, err := validator.GetSupportedValidators(n.Identifier)
	if err != nil {
		return validator.ValidatorForNetwork{}, errors.Wrap(err, 0)
	}

	var validators []validator.ValidatorForNetwork
	if len(delegations) == 0 {
		validators = supportedValidators
	} else {
		for _, v := range supportedValidators {
			for _, d := range delegations {
				if d.ValidatorAddress == v.ValidatorAddress {
					validators = append(validators, v)
				}
			}
		}
	}

	if len(validators) == 0 {
		return validator.ValidatorForNetwork{}, errors.New("no relevant validators found (is the address staking with a REStake-enabled validator on this network?)")
	}

	validatorNameList := make([]string, len(validators))
	for i, v := range validators {
		validatorNameList[i] = v.Name
	}
	var validatorChoiceIndex int
	validatorPrompt := &survey.Select{
		Message: "Choose validator",
		Options: validatorNameList,
	}
	err = survey.AskOne(validatorPrompt, &validatorChoiceIndex)
	if err != nil {
		return validator.ValidatorForNetwork{}, errors.Wrap(err, 0)
	}
	return validators[validatorChoiceIndex], nil
}
