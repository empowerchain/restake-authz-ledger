package validator

import (
	"encoding/json"
	"github.com/go-errors/errors"
	"net/http"
)

type ValidatorForNetwork struct {
	Path             string
	Name             string
	Identity         string
	RestakeAddress   string
	ValidatorAddress string
}

type ValidatorsResponseJSON struct {
	Validators []ValidatorJSON `json:"validators"`
}

type ValidatorJSON struct {
	Path     string      `json:"path"`
	Name     string      `json:"name"`
	Identity string      `json:"identity"`
	Chains   []ChainJSON `json:"chains"`
}

type ChainJSON struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Restake string `json:"restake"`
}

func GetSupportedValidators(network string) (supportedValidators []ValidatorForNetwork, err error) {
	res, err := http.Get("https://validators.cosmos.directory")
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.Errorf("Getting validators failed with a %d code", res.StatusCode)
	}

	validatorsRes := ValidatorsResponseJSON{}
	err = json.NewDecoder(res.Body).Decode(&validatorsRes)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	for _, v := range validatorsRes.Validators {
		for _, c := range v.Chains {
			if c.Name == network {
				supportedValidators = append(supportedValidators, ValidatorForNetwork{
					Path:             v.Path,
					Name:             v.Name,
					Identity:         v.Identity,
					RestakeAddress:   c.Restake,
					ValidatorAddress: c.Address,
				})
			}
		}
	}

	return supportedValidators, nil
}
